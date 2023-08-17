package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

var ministart time.Time

func ZeroRegisters(ram map[uint32](uint32)) {
	for i := uint32(0xC0000000); i < 0xC0000000+36*4; i += 4 {
		WriteRam(ram, i, 0)
	}
}

func LoadData(dat []byte, ram map[uint32](uint32), base uint32) {
	for i := 0; i < len(dat); i += 4 {
		value := binary.BigEndian.Uint32(dat[i : i+4])
		if value != 0 {
			ram[base+uint32(i)] = value
		}
	}
}

func LoadMappedFile(fn string, ram map[uint32](uint32), base uint32) {
	dat, err := ioutil.ReadFile(fn)
	check(err)
	LoadData(dat, ram, base)
}

func Uint32ToBytes(x uint32, isBigEndian bool) []byte {
    bytesBuffer := bytes.NewBuffer([]byte{})
	if isBigEndian {
		binary.Write(bytesBuffer, binary.BigEndian, x)
	} else {
		binary.Write(bytesBuffer, binary.LittleEndian, x)
	}
    return bytesBuffer.Bytes()
}

func SaveOutput(outputPath string, ram map[uint32](uint32)) error {
	output := make([]byte, 0)
	size := ram[0x32000000]
	for i := uint32(0); i < size; i+=4 {
		output = append(output, Uint32ToBytes(ram[0x32000004+i], false)...)
	}
	fout, err := os.Create(outputPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer fout.Close()
	_, err = fout.Write(output)
	return err  
}