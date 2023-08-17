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
	}

	return params
}

func Run() {
	params := ParseParams()
	RunWithParams(params)
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
		nodeFile, nodeCount, err := LayerRun(basedir + "/data", id, modelName)
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

func LayerRun(basedir string, nodeID int, modelName string) (string, int, error) {
	var envBytes []byte
	var err error
	var nodeCount int

	if modelName == "MNIST" {
		envBytes, nodeCount, err = MNIST(nodeID)
	} else { // if modelName == "LLAMA"
		envBytes, nodeCount, err = LLAMA(nodeID)
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
			if step == target {
				// done
				mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
			}
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

	mu.Start(0, 0x5ead0004)


	if reachFinalState {
		fmt.Printf("reach the final state, total step: %d, target: %d\n", lastStep, target)
		WriteCheckpointWithNodeID(ram, fmt.Sprintf("%s/checkpoint_%d_%d.json", basedir, nodeID, lastStep), lastStep, nodeID, nodeCount)
	}

	if target == -1 {

		fmt.Println("lastStep: ", lastStep)
		WriteCheckpointWithNodeID(ram, fmt.Sprintf("%s/checkpoint_%d_final.json", basedir, nodeID), lastStep, nodeID, nodeCount)

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
			if step == target {
				// done
				mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
			}
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