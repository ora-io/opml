package vm

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func initTest() {
	Preimages = make(map[common.Hash][]byte)
	steps = 0
	heap_start = 0
}

func TestVM(t *testing.T){
	initTest()
	params := &Params{
		Target: 2,
		ProgramPath: MIPS_PROGRAM,
		Basedir: "/tmp/cannon",
		ModelName: "MNIST",
		LastLayer: false,
		NodeID: 0,
	}
	RunWithParams(params)
}

func TestVM1(t *testing.T){
	initTest()
	params := &Params{
		Target: 0,
		ProgramPath: MIPS_PROGRAM,
		Basedir: "/tmp/cannon",
		ModelName: "LLAMA",
		LastLayer: false,
		NodeID: 0,
	}
	RunWithParams(params)
}

func TestVM2(t *testing.T){
	initTest()
	params := &Params{
		Target: -1,
		ProgramPath: MIPS_PROGRAM,
		Basedir: "/tmp/cannon",
		ModelName: "MNIST",
		LastLayer: true,
		InputPath: "/tmp/cannon/data/node_2",
		NodeID: 2,
	}
	RunWithParams(params)
}

func TestVM3(t *testing.T){
	initTest()
	params := &Params{
		Target: -1,
		ProgramPath: "../../mlgo/examples/mnist_mips/mlgo.bin",
		Basedir: "/tmp/cannon",
		ModelPath: "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin",
		InputPath: "../../mlgo/examples/mnist/models/mnist/input_7",
		MIPSVMCompatible: true,
	}
	RunWithParams(params)
}

func TestMPVM(t *testing.T) {
	initTest()
	params := &MPParams{
		ProgramPath: MIPS_PROGRAM,
		Basedir: "/tmp/cannon",
		ModelPath: "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin",
		InputPath: "../../mlgo/examples/mnist/models/mnist/input_7",
		ModelName: "MNIST",

		CurPhase: 0,
		TotalPhase: 2,
		Checkpoints: []int{2},
		StepCount: make([]int, 2),
	}
	RunWithMPParams(params)
}

func TestMPVMLast(t *testing.T) {
	initTest()
	params := &MPParams{
		ProgramPath: MIPS_PROGRAM,
		Basedir: "/tmp/cannon",
		// ModelPath: "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin",
		InputPath: "/tmp/cannon/data/[2].dat",
		ModelName: "MNIST",

		CurPhase: 0,
		TotalPhase: 2,
		Checkpoints: []int{2,0},
		StepCount: make([]int, 2),
	}
	RunWithMPParams(params)
}