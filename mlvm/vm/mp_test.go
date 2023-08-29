package vm

import (
	"testing"

)

func TestExecCommand(t *testing.T) {
	params := &MPParams{
		Basedir: "/tmp/cannon",
		TotalPhase: 3,
		Checkpoints: []int{1,2,3},
		StepCount: []int{10,20,30},
		ExecCommand: "python ../scripts/server.py",
		ExecOutputDir: "checkpoint",
	}
	MPExecCommand(params)
}