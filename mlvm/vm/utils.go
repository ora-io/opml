package vm

import (
	"encoding/binary"
	"io/ioutil"
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