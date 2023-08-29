package vm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/common"
)

func MPExecCommand(params *MPParams) error {
	outputPath := params.Basedir + "/" + params.ExecOutputDir + "/" + IntList2String(params.Checkpoints) + ".json"
	scriptPath := params.Basedir +  "/" + params.ExecOutputDir + "/" + "command.sh"
	
	fn, err := os.Create(scriptPath)
	defer fn.Close()

	if err != nil {
		return err
	}

	if params.ExecWorkDir != ""{
		fn.WriteString(fmt.Sprintf("cd %s\n", params.ExecWorkDir))
	}
	

	if params.ExecCommand != "" {
		execCommand := fmt.Sprintf("%s --mp_checkpoints %s --mp_stepCount %s --mp_execOutputPath %s", params.ExecCommand, IntList2String(params.Checkpoints), IntList2String(params.StepCount), outputPath)
		fn.WriteString(execCommand + "\n")
	}

	result, err := exec.Command("sh", scriptPath).Output()
	if err != nil {
		fmt.Println("exec command error: ", err)
		return err
	}
	fmt.Println("exec command output: ", string(result))

	return nil
}

func MockWriteCheckpoint(data []byte, fn string, checkpoints []int, stepCount []int) {
	trieroot := common.BytesToHash(data)
	dat := MPTrieToJson(trieroot, checkpoints, stepCount)
	fmt.Printf("writing %s len %d with root %s\n", fn, len(dat), trieroot)
	ioutil.WriteFile(fn, dat, 0644)
}

// TODO: we should generate the checkpoint in the script, since we need to know about the step count
// func MPGenerateCheckpoint(params *MPParams) error {
// 	dataPath := params.Basedir + "/" + params.ExecOutputDir + "/" + IntList2String(params.Checkpoints) + ".dat"
// 	checkpointPath := fmt.Sprintf("%s/checkpoint/%s.json", params.Basedir, IntList2String(params.Checkpoints))
// 	dataBytes, err := ioutil.ReadFile(dataPath)
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}
// 	return nil
// }

func MPRun(params *MPParams) error {
	// exec the command first to generate output
	return MPExecCommand(params)
}