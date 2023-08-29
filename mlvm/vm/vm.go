package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

func WriteCheckpoint(ram map[uint32](uint32), fn string, step int) {
	trieroot := RamToTrie(ram)
	dat := TrieToJson(trieroot, step)
	fmt.Printf("writing %s len %d with root %s\n", fn, len(dat), trieroot)
	ioutil.WriteFile(fn, dat, 0644)
}

func WriteCheckpointWithNodeID(ram map[uint32](uint32), fn string, step int, nodeID int, nodeCount int) {
	trieroot := RamToTrie(ram)
	dat := TrieToJsonWithNodeID(trieroot, step, nodeID, nodeCount)
	fmt.Printf("writing %s len %d with root %s\n", fn, len(dat), trieroot)
	ioutil.WriteFile(fn, dat, 0644)
}

func MPWriteCheckpoint(ram map[uint32](uint32), fn string, checkpoints []int, stepCount []int) {
	trieroot := RamToTrie(ram)
	dat := MPTrieToJson(trieroot, checkpoints, stepCount)
	fmt.Printf("writing %s len %d with root %s\n", fn, len(dat), trieroot)
	ioutil.WriteFile(fn, dat, 0644)
}


// memory layout in MIPS
const (
	INPUT_ADDR = 0x31000000
	OUTPUT_ADDR = 0x32000000
	MODEL_ADDR = 0x33000000
	MAGIC_ADDR = 0x30000800
)

const (
	MIPS_PROGRAM = "../../mlgo/ml_mips/ml_mips.bin"
)

const (
	READ_FROM_BIDENDIAN = true
	OUTPUT_TO_BIDENDIAN = true
)

func IntToBytes(n int) []byte {
    x := int32(n)
    bytesBuffer := bytes.NewBuffer([]byte{})
	if READ_FROM_BIDENDIAN{
		binary.Write(bytesBuffer, binary.BigEndian, x)
	} else {
		binary.Write(bytesBuffer, binary.LittleEndian, x)
	}
    
    return bytesBuffer.Bytes()
}

func LoadModel(mu uc.Unicorn, file string, ram map[uint32](uint32)) {
	modelBytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	modelSize := len(modelBytes)
	fmt.Println("modelSize: ", modelSize)
	rawSize := IntToBytes(modelSize)
	fmt.Println("rawSize: ", rawSize)
	LoadBytesToUnicorn(mu, rawSize, ram, MODEL_ADDR)
	LoadBytesToUnicorn(mu, modelBytes, ram, MODEL_ADDR + 4)
}

func LoadInputData(mu uc.Unicorn, file string, ram map[uint32](uint32)) error {
	// load a random test digit
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if len(buf) >= 10 * 1024 * 1024 {
		fmt.Println("data too large")
		return errors.New("data too large")
	}
	//buf is the data
	inputSize := len(buf)
	LoadBytesToUnicorn(mu, IntToBytes(inputSize), ram, INPUT_ADDR)
	LoadBytesToUnicorn(mu, buf, ram, INPUT_ADDR + 4)
	
	return nil
}

type Params struct {
	Target int
	ProgramPath string
	ModelPath string
	InputPath string
	Basedir string
	OutputGolden bool

	CurLayer int
	LastLayer bool
	ModelName string
	NodeID int

	MIPSVMCompatible bool
	Prompt string
}

type MPParams struct {
	ProgramPath string `json:"programPath"`
	ModelPath string 	`json:"modelPath"`
	InputPath string	`json:"inputPath"`
	Basedir string		`json:"basedir"`
	ModelName string	`json:"modelName"`

	CurPhase int		`json:"curPhase"`
	TotalPhase int		`json:"totalPhase"`
	Checkpoints []int	`json:"checkpoints"`
	StepCount []int		`json:"stepCounts"`

	// optional
	Prompt string		`json:"prompt"`
	// for exec
	ExecCommand string	`json:"execCommand"` 
	ExecOutputDir string `json:"execOutputDir"`
	ExecWorkDir string	`json:"execWorkDir"`
}

func ParseMPParams() (*MPParams, error) {

	var mp bool
	flag.BoolVar(&mp, "mp", false, "enable mp mode")

	var programPath string
	var modelPath string
	var inputPath string
	var basedir string
	var modelName string

	var curPhase int
	var totalPhase int 
	var checkpoints_json string
	var stepcount_json string

	var prompt string
	// for exec
	var execCommand string
	var execOutputDir string
	var execWorkDir string

	defaultBasedir := os.Getenv("BASEDIR")
	if len(defaultBasedir) == 0 {
		defaultBasedir = "/tmp/cannon"
	}
	flag.StringVar(&basedir, "basedir", defaultBasedir, "Directory to read inputs, write outputs, and cache preimage oracle data.")
	flag.StringVar(&programPath, "program", MIPS_PROGRAM, "Path to binary file containing the program to run")
	flag.StringVar(&modelPath, "model", "", "Path to binary file containing the AI model")
	flag.StringVar(&inputPath, "data", "", "Path to binary file containing the input of AI model")
	flag.StringVar(&modelName, "modelName", "MNIST", "run MNIST or LLAMA")
	flag.IntVar(&curPhase, "curPhase", 0, "The current phase")
	flag.IntVar(&totalPhase, "totalPhase", 0, "The total number of phases")
	flag.StringVar(&checkpoints_json, "checkpoints", "[]", "The checkpoint includes the phase info, e.g., [1, 2, 3] means Phase-0 is at stage 1, Phase-1 is at stage 2, Phase-2 is at stage 3")
	// [0] means the golden of the initial stage, [1] means the image at the begin 1st stage
	flag.StringVar(&stepcount_json, "stepCount", "[]", "the total number of steps at the current phases")

	flag.StringVar(&prompt, "prompt", "How to combine AI and blockchain?", "prompt for LLaMA")

	// for exec
	flag.StringVar(&execCommand, "execCommand", "", "the exec command for generating checkpoints")
	flag.StringVar(&execOutputDir, "execOutputDir", "checkpoint", "save the output data for generating checkpoints, basedir/execOutputDir")
	flag.StringVar(&execWorkDir, "execWorkDir", "", "working dir, will cd execWorkDir first and then exec the command")

	flag.Parse()

	checkpoints, err := Strings2IntList(checkpoints_json)
	if err != nil {
		fmt.Println("ParseMPParams error: ", err)
		return nil, err
	}
	if err := ValidateCheckpoints(checkpoints, totalPhase, curPhase); err != nil {
		fmt.Println("ParseMPParams error: ", err)
		return nil, err		
	}

	stepCount, err := Strings2IntList(stepcount_json)
	if err != nil {
		fmt.Println("ParseMPParams error: ", err)
		return nil, err
	}
	padding := totalPhase - len(stepCount)
	for i := 0; i < padding; i++ {
		stepCount = append(stepCount, 0)
	}
	if len(stepCount) > totalPhase {
		stepCount = stepCount[:totalPhase]
	}
	// padding with zero, such that len(stepCount) == totalPhase

	params := &MPParams{
		ProgramPath: programPath,
		ModelPath: modelPath,
		InputPath: inputPath,
		Basedir: basedir,
		ModelName: modelName,

		CurPhase: curPhase,
		TotalPhase: totalPhase,
		Checkpoints: checkpoints,
		StepCount: stepCount,

		Prompt: prompt,

		ExecCommand: execCommand,
		ExecOutputDir: execOutputDir,
		ExecWorkDir: execWorkDir,
	}
	return params, nil
}

func ParseParams() *Params {
	var target int
	var programPath string
	var modelPath string
	var inputPath string
	var basedir string
	var outputGolden bool

	var curLayer int
	var lastLayer bool
	var modelName string
	var nodeID int

	var mipsVMCompatible bool
	var prompt string

	defaultBasedir := os.Getenv("BASEDIR")
	if len(defaultBasedir) == 0 {
		defaultBasedir = "/tmp/cannon"
	}
	flag.StringVar(&basedir, "basedir", defaultBasedir, "Directory to read inputs, write outputs, and cache preimage oracle data.")
	flag.IntVar(&target, "target", -1, "Target number of instructions to execute in the trace. If < 0 will execute until termination")
	flag.StringVar(&programPath, "program", MIPS_PROGRAM, "Path to binary file containing the program to run")
	flag.StringVar(&modelPath, "model", "", "Path to binary file containing the AI model")
	flag.StringVar(&inputPath, "data", "", "Path to binary file containing the input of AI model")
	flag.BoolVar(&outputGolden, "outputGolden", false, "Do not read any inputs and instead produce a snapshot of the state prior to execution. Written to <basedir>/golden.json")

	flag.BoolVar(&lastLayer, "lastLayer", false, "In the lastLayer, we run computation in VM")
	flag.IntVar(&curLayer, "curLayer", 0, "The current layer")
	flag.StringVar(&modelName, "modelName", "MNIST", "run MNIST or LLAMA")
	flag.IntVar(&nodeID, "nodeID", 0, "The current nodeID")
	
	flag.BoolVar(&mipsVMCompatible, "mipsVMCompatible", false, "compatible for MIPS VM")
	flag.StringVar(&prompt, "prompt", "How to combine AI and blockchain?", "prompt for LLaMA")
	flag.Parse()

	params := &Params{
		Target: target,
		ProgramPath: programPath,
		ModelPath: modelPath,
		InputPath: inputPath,
		Basedir: basedir,
		OutputGolden: outputGolden,
		CurLayer: curLayer,
		LastLayer: lastLayer,
		ModelName: modelName,
		NodeID: nodeID,
		MIPSVMCompatible: mipsVMCompatible,
		Prompt: prompt,
	}

	return params
}

func Run() {
	fmt.Println("start!!!")
	mpMode := os.Args[1] == "--mp"
	// flag.BoolVar(&mpMode, "mpMode", false, "enable mpMode")
	// fmt.Println("mpMode: ", mpMode)
	// fmt.Println(os.Args)
	if mpMode {
		fmt.Println("run in mp mode")
		mpParams, err := ParseMPParams()
		if err != nil {
			fmt.Println("ParseMPParams error: ", err)
			return 
		}
		RunWithMPParams(mpParams)
	} else {
		params := ParseParams()
		RunWithParams(params)
	}

}

func RunWithMPParams(params *MPParams) {
	// single-phase opml, equal to params.MIPSVMCompatible
	if params.TotalPhase == 1 {
		outputGolden := (params.Checkpoints[0] == 0)
		MIPSRunCompatible(params.Basedir, params.Checkpoints[0], params.ProgramPath, params.ModelPath, params.InputPath, outputGolden)
		return 
	}

	// multi-phase opml
	if len(params.Checkpoints) == params.TotalPhase {
		// the last phase, we should run it in VM
		// assume we are bisect to one node on computation graph now
		MPMIPSRun(params) 
		return 
	} else if len(params.Checkpoints) == params.TotalPhase - 1 {
		// the penultimate, dispute on the computation graph
		nodeFile, err := MPGraphRun(params)
		if err != nil {
			fmt.Println("layer run error: ", err, "nodeFile: ", nodeFile)
			return
		}
		params.Checkpoints = append(params.Checkpoints, 0)
		MPMIPSRun(params) // init golden for the last phase
		// copy the file as the checkpoint output, should modify it later
		err = CopyFile(fmt.Sprintf("%s/checkpoint/%s.json", params.Basedir, IntList2String(params.Checkpoints)), fmt.Sprintf("%s/checkpoint/%s.json",  params.Basedir, IntList2String(params.Checkpoints[:len(params.Checkpoints)-1])))
		fmt.Println("copyFile error: ", err)
		return 
	} else {
		// multi-phase opml
		MPRun(params)
	}
}

func RunWithParams(params *Params) {

	target := params.Target
	programPath := params.ProgramPath
	modelPath := params.ModelPath
	inputPath := params.InputPath
	basedir := params.Basedir
	outputGolden := params.OutputGolden
	// curLayer := params.CurLayer
	lastLayer := params.LastLayer
	modelName := params.ModelName
	nodeID := params.NodeID

	if params.MIPSVMCompatible {
		MIPSRunCompatible(basedir, target, programPath, modelPath, inputPath, outputGolden)
		return
	}

	if !lastLayer {
		id := target
		nodeFile, nodeCount, err := LayerRun(basedir + "/data", id, modelName, params)
		if err != nil {
			fmt.Println("layer run error: ", err)
			return
		}
		MIPSRun(basedir + "/checkpoint", 0, id, programPath, nodeFile, true, nodeCount)
	} else {
		// the lastLayer
		MIPSRun(basedir + "/checkpoint", target, nodeID, programPath, inputPath, outputGolden, 0)
	}
	

	// step 2 (optional), validate each 1 million chunk in EVM

	// step 3 (super optional) validate each 1 million chunk on chain

	//RunWithRam(ram, steps, debug, nil)

}

func LayerRun(basedir string, nodeID int, modelName string, params *Params) (string, int, error) {
	var envBytes []byte
	var err error
	var nodeCount int

	if modelName == "MNIST" {
		envBytes, nodeCount, err = MNIST(nodeID, params.ModelPath, params.InputPath)
	} else { // if modelName == "LLAMA"
		envBytes, nodeCount, err = LLAMA(nodeID, params.ModelPath, params.Prompt)
	}

	if err != nil {
		fmt.Println("Layer run error: ", err)
		return "", nodeCount, err
	}

	fileName := fmt.Sprintf("%s/node_%d", basedir, nodeID)
	err = saveDataToFile(envBytes, fileName)

	if err != nil {
		fmt.Println("Save data error: ", err)
		return fileName, nodeCount, err
	}

	return fileName, nodeCount, nil
}

func MPGraphRun(params *MPParams) (string,  error) {
	var envBytes []byte
	var err error
	var nodeCount int

	nodeID := params.Checkpoints[params.TotalPhase - 2]

	if params.ModelName == "MNIST" {
		envBytes, nodeCount, err = MNIST(nodeID, params.ModelPath, params.InputPath)
	} else if params.ModelName == "LLAMA" { // if modelName == "LLAMA"
		envBytes, nodeCount, err = LLAMA(nodeID, params.ModelPath, params.Prompt)
	} else {
		envBytes, nodeCount, err = MNIST(nodeID, params.ModelPath, params.InputPath)
	}

	// update
	params.StepCount[params.TotalPhase - 2] = nodeCount

	if err != nil {
		fmt.Println("Layer run error: ", err)
		return "", err
	}

	fileName := fmt.Sprintf("%s/data/%s.dat", params.Basedir, IntList2String(params.Checkpoints))
	err = saveDataToFile(envBytes, fileName)

	if err != nil {
		fmt.Println("Save data error: ", err)
		return fileName, err
	}

	return fileName, nil
}



func saveDataToFile(data []byte, filename string) error {
	fout, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer fout.Close()
	_, err = fout.Write(data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func MIPSRun(basedir string, target int, nodeID int, programPath string, inputPath string, outputGolden bool, nodeCount int) {
	regfault := -1
	regfault_str, regfault_valid := os.LookupEnv("REGFAULT")
	if regfault_valid {
		regfault, _ = strconv.Atoi(regfault_str)
	}

	// step 1, generate the checkpoints every million steps using unicorn
	ram := make(map[uint32](uint32))

	lastStep := 1
	reachFinalState := true // if the target >= total step, the targt will not be saved

	mu := GetHookedUnicorn(basedir, ram, func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		if step == regfault {
			fmt.Printf("regfault at step %d\n", step)
			mu.RegWrite(uc.MIPS_REG_V0, 0xbabababa)
		}
		if step == target {
			reachFinalState = false
			SyncRegs(mu, ram)
			fn := fmt.Sprintf("%s/checkpoint_%d_%d.json", basedir, nodeID, step)
			WriteCheckpointWithNodeID(ram, fn, step, nodeID, nodeCount)
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
		lastStep = step + 1
	})

	ZeroRegisters(ram)
	// not ready for golden yet
	LoadMappedFileUnicorn(mu, programPath, ram, 0)
	// load input
	if inputPath != "" {
		LoadInputData(mu, inputPath, ram)
	}
	
	if outputGolden {
		WriteCheckpointWithNodeID(ram, fmt.Sprintf("%s/%d_golden.json", basedir, nodeID), -1, nodeID, nodeCount)
		fmt.Println("Writing golden snapshot and exiting early without execution")
		return 
	}

	// do not need if we just run pure computation task
	// LoadMappedFileUnicorn(mu, fmt.Sprintf("%s/input", basedir), ram, 0x30000000)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	if reachFinalState {
		fmt.Printf("reach the final state, total step: %d, target: %d\n", lastStep, target)
		WriteCheckpointWithNodeID(ram, fmt.Sprintf("%s/checkpoint_%d_%d.json", basedir, nodeID, lastStep), lastStep, nodeID, nodeCount)
	}

	if target == -1 {

		fmt.Println("lastStep: ", lastStep)
		WriteCheckpointWithNodeID(ram, fmt.Sprintf("%s/checkpoint_%d_final.json", basedir, nodeID), lastStep, nodeID, nodeCount)

	}
}

func MPMIPSRun(params *MPParams) {
	regfault := -1
	regfault_str, regfault_valid := os.LookupEnv("REGFAULT")
	if regfault_valid {
		regfault, _ = strconv.Atoi(regfault_str)
	}

	// step 1, generate the checkpoints every million steps using unicorn
	ram := make(map[uint32](uint32))

	lastStep := 1
	reachFinalState := true // if the target >= total step, the targt will not be saved

	target := params.Checkpoints[params.TotalPhase - 1]

	mu := GetHookedUnicorn(params.Basedir, ram, func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		if step == regfault {
			fmt.Printf("regfault at step %d\n", step)
			mu.RegWrite(uc.MIPS_REG_V0, 0xbabababa)
		}
		if step == target {
			reachFinalState = false
			SyncRegs(mu, ram)
			fn := fmt.Sprintf("%s/checkpoint/%s.json", params.Basedir, IntList2String(params.Checkpoints))
			MPWriteCheckpoint(ram, fn, params.Checkpoints, params.StepCount)
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
		lastStep = step + 1
	})

	ZeroRegisters(ram)
	// not ready for golden yet
	LoadMappedFileUnicorn(mu, params.ProgramPath, ram, 0)
	// load input
	if params.InputPath != "" {
		LoadInputData(mu, params.InputPath, ram)
	}
	
	// outputGolden := (params.Checkpoints[len(params.Checkpoints) - 1] == 0)
	// if outputGolden {
	// 	fn := fmt.Sprintf("%s/%d_golden.json", params.Basedir, nodeID)
	// 	MPWriteCheckpoint(ram, fn, params.Checkpoints, params.StepCount)
	// 	fmt.Println("Writing golden snapshot and exiting early without execution")
	// 	return 
	// }

	// do not need if we just run pure computation task
	// LoadMappedFileUnicorn(mu, fmt.Sprintf("%s/input", basedir), ram, 0x30000000)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)


	if reachFinalState {
		fmt.Printf("reach the final state, total step: %d, target: %d\n", lastStep, target)
		// update
		params.Checkpoints[params.TotalPhase - 1] = lastStep
		params.StepCount[params.TotalPhase - 1] = lastStep

		fn := fmt.Sprintf("%s/checkpoint/%s.json", params.Basedir, IntList2String(params.Checkpoints))
		MPWriteCheckpoint(ram, fn, params.Checkpoints, params.StepCount)
	}


	// only for the name
	targetCheckpoints := make([]int, len(params.Checkpoints))
	copy(targetCheckpoints, params.Checkpoints)
	targetCheckpoints[len(targetCheckpoints) - 1] = -1 // replace the last one with

	if target == -1 {
		fmt.Println("lastStep: ", lastStep)
		// update
		params.Checkpoints[params.TotalPhase - 1] = lastStep
		params.StepCount[params.TotalPhase - 1] = lastStep
		fn := fmt.Sprintf("%s/checkpoint/%s.json", params.Basedir, IntList2String(targetCheckpoints))
		MPWriteCheckpoint(ram, fn, params.Checkpoints, params.StepCount)
	}
}

func MIPSRunCompatible(basedir string, target int, programPath string, modelPath string, inputPath string, outputGolden bool) {
	regfault := -1
	regfault_str, regfault_valid := os.LookupEnv("REGFAULT")
	if regfault_valid {
		regfault, _ = strconv.Atoi(regfault_str)
	}

	// step 1, generate the checkpoints every million steps using unicorn
	ram := make(map[uint32](uint32))

	lastStep := 1
	reachFinalState := true // if the target >= total step, the targt will not be saved

	mu := GetHookedUnicorn(basedir, ram, func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		if step == regfault {
			fmt.Printf("regfault at step %d\n", step)
			mu.RegWrite(uc.MIPS_REG_V0, 0xbabababa)
		}
		if step == target {
			reachFinalState = false
			SyncRegs(mu, ram)
			fn := fmt.Sprintf("%s/checkpoint_%d.json", basedir, step)
			WriteCheckpoint(ram, fn, step)
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
		lastStep = step + 1
	})

	ZeroRegisters(ram)
	// not ready for golden yet
	LoadMappedFileUnicorn(mu, programPath, ram, 0)
	// load input
	if inputPath != "" {
		LoadInputData(mu, inputPath, ram)
	}
	LoadModel(mu, modelPath, ram)
	
	
	if outputGolden {
		WriteCheckpoint(ram, fmt.Sprintf("%s/golden.json", basedir), -1)
		fmt.Println("Writing golden snapshot and exiting early without execution")
		return 
	}

	// do not need if we just run pure computation task
	// LoadMappedFileUnicorn(mu, fmt.Sprintf("%s/input", basedir), ram, 0x30000000)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	if reachFinalState {
		fmt.Printf("reach the final state, total step: %d, target: %d\n", lastStep, target)
		WriteCheckpoint(ram, fmt.Sprintf("%s/checkpoint_%d.json", basedir, lastStep), lastStep)
	}

	if target == -1 {

		fmt.Println("lastStep: ", lastStep)
		WriteCheckpoint(ram, fmt.Sprintf("%s/checkpoint_final.json", basedir), lastStep)
		fmt.Printf("PC: %x\n", ram[0xC0000080])
		SaveOutput(fmt.Sprintf("%s/output", basedir), ram)
	}
}