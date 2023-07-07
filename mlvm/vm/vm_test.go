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