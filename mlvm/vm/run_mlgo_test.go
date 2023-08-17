package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)



func LoadMNISTData(mu uc.Unicorn, file string, ram map[uint32](uint32)) error {
	// load a random test digit
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// fin, err := os.Open(file)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// buf := make([]byte, 784)
	// if count, err := fin.Read(buf); err != nil || count != int(len(buf)) {
	// 	fmt.Println(err, count)
	// 	return err
	// }
	
	// // render the digit in ASCII
	// for row := 0; row < 28; row++{
	// 	for col := 0; col < 28; col++ {
	// 		var c string
	// 		if buf[row*28 + col] > 230 {
	// 			c = "*"
	// 		} else {
	// 			c = "_"
	// 		}
	// 		fmt.Printf(c)
	// 	}
	// 	fmt.Println("")
	// }
	// fmt.Println("")

	//buf is the data
	inputSize := len(buf)
	LoadBytesToUnicorn(mu, IntToBytes(inputSize), ram, INPUT_ADDR)
	LoadBytesToUnicorn(mu, buf, ram, INPUT_ADDR + 4)
	
	return nil
}

// testing
func TestMLGo_MNIST(t *testing.T){
	fn := "../../mlgo/mlgo.bin"
	fn = "../../mlgo/examples/mnist_mips/mlgo.bin"
	// fn = "../../Rollup_DL/mipigo/test/test2.bin" //for testing
	modelLittleEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	modelBigEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/t10k-images.idx3-ubyte"
	dataFile = "../../mlgo/examples/mnist/models/mnist/input_7"
	steps := 1000000000
	// steps = 12066041

	// reachFinalState := true

	modelFile := modelLittleEndianFile
	if READ_FROM_BIDENDIAN {
		modelFile = modelBigEndianFile
	}

	totalSteps := 0;

	ram := make(map[uint32](uint32))

	callback := func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		totalSteps += 1;
		// sync at each step is very slow
		// SyncRegs(mu, ram) 
		if step%10000000 == 0 {
			steps_per_sec := float64(step) * 1e9 / float64(time.Now().Sub(ministart).Nanoseconds())
			fmt.Printf("%10d pc: %x steps per s %f ram entries %d\n", step, ram[0xc0000080], steps_per_sec, len(ram))
		}
		// halt at steps
		if step == steps {
			fmt.Println("what what what ? final!")
			// reachFinalState = false
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
	}

	mu := GetHookedUnicorn("", ram, callback)
	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadModel(mu, modelFile, ram)
	LoadMNISTData(mu, dataFile, ram)
	
	// initial checkpoint
	// WriteCheckpoint(ram, "/tmp/cannon/golden.json", 0)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	// final checkpoint
	// if reachFinalState {
	// 	WriteCheckpoint(ram, fmt.Sprintf("/tmp/cannon/checkpoint_%d.json", totalSteps), totalSteps)
	// }
	WriteCheckpoint(ram, "/tmp/cannon/checkpoint_final.json", totalSteps)

	SyncRegs(mu, ram)

	fmt.Println("ram[0x32000000]: ", ram[0x32000000])
	fmt.Println("ram[0x32000004]: ", ram[0x32000004])
	fmt.Printf("PC: %x\n", ram[0xC0000080])

	fmt.Println("total steps: ", totalSteps)

	SaveOutput("/tmp/cannon/output", ram)
}


// test on node
func TestMLGo_MNIST_Node(t *testing.T){
	fn := "../../mlgo/ml_mips/ml_mips.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/node_5"
	dataFile = "../../mlgo/examples/llama/data/node_1253"
	steps := 1000000000
	// steps = 12066041


	totalSteps := 0;

	ram := make(map[uint32](uint32))

	callback := func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		totalSteps += 1;
		// sync at each step is very slow
		// SyncRegs(mu, ram) 
		if step%10000000 == 0 {
			steps_per_sec := float64(step) * 1e9 / float64(time.Now().Sub(ministart).Nanoseconds())
			fmt.Printf("%10d pc: %x steps per s %f ram entries %d\n", step, ram[0xc0000080], steps_per_sec, len(ram))
		}
		// halt at steps
		if step == steps {
			fmt.Println("what what what ? final!")
			// reachFinalState = false
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
	}

	mu := GetHookedUnicorn("", ram, callback)
	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadInputData(mu, dataFile, ram)
	
	// initial checkpoint
	// WriteCheckpoint(ram, "/tmp/cannon/golden.json", 0)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	// final checkpoint
	// if reachFinalState {
	// 	WriteCheckpoint(ram, fmt.Sprintf("/tmp/cannon/checkpoint_%d.json", totalSteps), totalSteps)
	// }
	WriteCheckpoint(ram, "/tmp/cannon/checkpoint_final.json", totalSteps)

	SyncRegs(mu, ram)

	fmt.Println("total steps: ", totalSteps)
}

func BytesToInt32(b []byte, isBigEndian bool) uint32 {
    bytesBuffer := bytes.NewBuffer(b)

    var x uint32
	if isBigEndian {
		binary.Read(bytesBuffer, binary.BigEndian, &x)
	} else {
		binary.Read(bytesBuffer, binary.LittleEndian, &x)
	}
    

    return x
}

// Faster design! we do not store 0 in ram and trie
func TestMLGo_MNIST2_Fast(t *testing.T){
	fn := "../../mlgo/mlgo.bin"
	// fn = "../../Rollup_DL/mipigo/test/test2.bin" //for testing
	modelLittleEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	modelBigEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-f32-big-endian.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/t10k-images.idx3-ubyte"
	dataFile = "../../mlgo/examples/mnist/models/mnist/input_7"
	steps := 1000000000
	calSteps := false
	// steps = 12066041

	// reachFinalState := true

	modelFile := modelLittleEndianFile
	if READ_FROM_BIDENDIAN {
		modelFile = modelBigEndianFile
	}

	totalSteps := 0;

	ram := make(map[uint32](uint32))


	mu := GetHookedUnicorn("", ram, nil)

	if calSteps {
		mu.HookAdd(uc.HOOK_CODE, func(mu uc.Unicorn, addr uint64, size uint32) {
			totalSteps += 1
		}, 0, 0x80000000)
	}


	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadModel(mu, modelFile, ram)
	LoadMNISTData(mu, dataFile, ram)

	SyncRegs(mu, ram)
	
	option := &uc.UcOptions{Timeout: 0, Count: uint64(steps)}
	mu.StartWithOptions(0, 0x5ead0004, option)

	// mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
	SyncRegs(mu, ram)

	memory, err := mu.MemRead(0,0x80000000-4)
	if err != nil {
		fmt.Println(err)
	}
	cnt := 0
	for i := 0; i < len(memory)-4; i += 4 {
		if memory[i] == 0 && memory[i+1] ==0 && memory[i+2] == 0 && memory[i+3] == 0 {
			continue
		}
		v := BytesToInt32(memory[i:i+4], true)
		if v != 0 {
			cnt += 1
			ram[uint32(i)] = v
		}
	}

	for k,v := range ram {
		if v == 0{
			delete(ram, k)
		}
	}

	fmt.Printf("cnt: %d, size of ram: %d\n", cnt, len(ram))
	// final checkpoint
	// if reachFinalState {
	// 	WriteCheckpoint(ram, fmt.Sprintf("/tmp/cannon/checkpoint_%d.json", totalSteps), totalSteps)
	// }
	WriteCheckpoint(ram, "/tmp/cannon/checkpoint_final_test.json", totalSteps)


	fmt.Println("ram[0x5ead0004]: ", ram[0x5ead0004])
	fmt.Println("ram[0x32000000]: ", ram[0x32000000])
	fmt.Println("ram[0x32000004]: ", ram[0x32000004])

	fmt.Println("total steps: ", totalSteps)

}


func TestMLGo_MNIST2(t *testing.T){
	fn := "../../mlgo/mlgo.bin"
	// fn = "../../Rollup_DL/mipigo/test/test2.bin" //for testing
	modelLittleEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	modelBigEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/t10k-images.idx3-ubyte"
	dataFile = "../../mlgo/examples/mnist/models/mnist/input_7"
	steps := 1000000000
	compare := true
	// steps = 12066041

	// reachFinalState := true

	modelFile := modelLittleEndianFile
	if READ_FROM_BIDENDIAN {
		modelFile = modelBigEndianFile
	}

	totalSteps := 0;

	ram := make(map[uint32](uint32))


	mu := GetHookedUnicorn("", ram, nil)

	mu.HookAdd(uc.HOOK_CODE, func(mu uc.Unicorn, addr uint64, size uint32) {
		totalSteps += 1
	}, 0, 0x80000000)

	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadModel(mu, modelFile, ram)
	LoadMNISTData(mu, dataFile, ram)

	SyncRegs(mu, ram)
	
	option := &uc.UcOptions{Timeout: 0, Count: uint64(steps)}
	mu.StartWithOptions(0, 0x5ead0004, option)

	// mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
	SyncRegs(mu, ram)

	memory, err := mu.MemRead(0,0x80000000-4)
	fmt.Println("memory len: ", len(memory))
	if err != nil {
		fmt.Println(err)
	}
	cnt := 0
	for i := 0; i < len(memory)-4; i += 4 {
		if memory[i] == 0 && memory[i+1] ==0 && memory[i+2] == 0 && memory[i+3] == 0 {
			continue
		}
		v := BytesToInt32(memory[i:i+4], true)
		if v != 0 {
			cnt += 1
			if ram_v, ok := ram[uint32(i)]; !ok || ram_v == v {
				ram[uint32(i)] = v
			} else  {
				fmt.Printf("ram address: %d updates from %d to %d\n", i, ram_v, v)
			}
			ram[uint32(i)] = v
		}
	}

	for k,v := range ram {
		if v == 0{
			delete(ram, k)
		}
	}

	fmt.Printf("cnt: %d, size of ram: %d\n", cnt, len(ram))
	// final checkpoint
	// if reachFinalState {
	// 	WriteCheckpoint(ram, fmt.Sprintf("/tmp/cannon/checkpoint_%d.json", totalSteps), totalSteps)
	// }
	WriteCheckpoint(ram, "/tmp/cannon/checkpoint_final_test.json", totalSteps)


	fmt.Println("ram[0x5ead0004]: ", ram[0x5ead0004])
	fmt.Println("ram[0x32000000]: ", ram[0x32000000])
	fmt.Println("ram[0x32000004]: ", ram[0x32000004])

	fmt.Println("total steps: ", totalSteps)

	// the difference happens because we do not delete the memory with 0 value!

	// compare
	if compare {
		steps = 0
		heap_start = 0
		old_ram := MLGo_MNIST2_helper()
		fmt.Printf("old ram len: %d, new ram len: %d\n", len(old_ram), len(ram))
	
		for k,v := range ram {
			old_v, ok := old_ram[k]
			if old_v != v || !ok{
				fmt.Printf("diff! In RAM! address: %x, new value: %d, old value: %d\n", k, v, old_v)
			}
		}
		for k,old_v := range old_ram {
			v, ok := ram[k]
			if old_v != v || !ok{
				fmt.Printf("diff! In OldRam! address: %x, new value: %d, old value: %d\n", k, v, old_v)
			}
		}
		if reflect.DeepEqual(ram, old_ram) {
			fmt.Println("equal")
		} else {
			fmt.Println("non-equal")
		}
	}
}

func MLGo_MNIST2_helper() map[uint32](uint32){
	fn := "../../mlgo/mlgo.bin"
	// fn = "../../Rollup_DL/mipigo/test/test2.bin" //for testing
	modelLittleEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	modelBigEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/t10k-images.idx3-ubyte"
	dataFile = "../../mlgo/examples/mnist/models/mnist/input_7"
	steps := 1000000000
	// steps = 12066041

	// reachFinalState := true

	modelFile := modelLittleEndianFile
	if READ_FROM_BIDENDIAN {
		modelFile = modelBigEndianFile
	}

	totalSteps := 0;

	ram := make(map[uint32](uint32))

	callback := func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		totalSteps += 1;
		// sync at each step is very slow
		// SyncRegs(mu, ram) 
		if step%10000000 == 0 {
			steps_per_sec := float64(step) * 1e9 / float64(time.Now().Sub(ministart).Nanoseconds())
			fmt.Printf("%10d pc: %x steps per s %f ram entries %d\n", step, ram[0xc0000080], steps_per_sec, len(ram))
		}
		// halt at steps
		if step == steps {
			fmt.Println("what what what ? final!")
			// reachFinalState = false
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
	}

	mu := GetHookedUnicorn("", ram, callback)
	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadModel(mu, modelFile, ram)
	LoadMNISTData(mu, dataFile, ram)
	
	// initial checkpoint
	// WriteCheckpoint(ram, "/tmp/cannon/golden.json", 0)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	for k,v := range ram {
		if v == 0{
			delete(ram, k)
		}
	}

	fmt.Println("ram[0x5ead0004]: ", ram[0x5ead0004])
	fmt.Println("ram[0x32000000]: ", ram[0x32000000])
	fmt.Println("ram[0x32000004]: ", ram[0x32000004])

	fmt.Println("total steps: ", totalSteps)

	return ram
}


func TestMLGo_MNIST3(t *testing.T){
	fn := "../../mlgo/mlgo.bin"
	// fn = "../../Rollup_DL/mipigo/test/test2.bin" //for testing
	modelLittleEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32.bin"
	modelBigEndianFile := "../../mlgo/examples/mnist/models/mnist/ggml-model-small-f32-big-endian.bin"
	dataFile := "../../mlgo/examples/mnist/models/mnist/t10k-images.idx3-ubyte"
	dataFile = "../../mlgo/examples/mnist/models/mnist/input_7"
	steps := 1000000000
	// steps = 12066041

	// reachFinalState := true

	modelFile := modelLittleEndianFile
	if READ_FROM_BIDENDIAN {
		modelFile = modelBigEndianFile
	}

	totalSteps := 0;

	ram := make(map[uint32](uint32))

	callback := func(step int, mu uc.Unicorn, ram map[uint32](uint32)) {
		totalSteps += 1;
		// sync at each step is very slow
		// SyncRegs(mu, ram) 
		if step%10000000 == 0 {
			steps_per_sec := float64(step) * 1e9 / float64(time.Now().Sub(ministart).Nanoseconds())
			fmt.Printf("%10d pc: %x steps per s %f ram entries %d\n", step, ram[0xc0000080], steps_per_sec, len(ram))
		}
		// halt at steps
		if step == steps {
			fmt.Println("what what what ? final!")
			// reachFinalState = false
			mu.RegWrite(uc.MIPS_REG_PC, 0x5ead0004)
		}
	}

	mu := GetHookedUnicorn("", ram, callback)
	// program 
	ZeroRegisters(ram)
	LoadMappedFileUnicorn(mu, fn, ram, 0)
	// load model and input
	LoadModel(mu, modelFile, ram)
	LoadMNISTData(mu, dataFile, ram)
	
	// initial checkpoint
	// WriteCheckpoint(ram, "/tmp/cannon/golden.json", 0)

	SyncRegs(mu, ram)
	mu.Start(0, 0x5ead0004)
	SyncRegs(mu, ram)

	// final checkpoint
	// if reachFinalState {
	// 	WriteCheckpoint(ram, fmt.Sprintf("/tmp/cannon/checkpoint_%d.json", totalSteps), totalSteps)
	// }
	
	// if we delete with 0 value, it should be the same 
	for k,v := range ram {
		if v == 0{
			delete(ram, k)
		}
	}

	WriteCheckpoint(ram, "/tmp/cannon/checkpoint_final.json", totalSteps)


	fmt.Println("ram[0x32000000]: ", ram[0x32000000])
	fmt.Println("ram[0x32000004]: ", ram[0x32000004])

	fmt.Println("total steps: ", totalSteps)
}
