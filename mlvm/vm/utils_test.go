package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestJSON(t *testing.T) {
	var list []int
	str := "[1, 2, 3]"
	err := json.Unmarshal([]byte(str), &list)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("list: ", list)
	new_str, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(new_str))
}


func TestXxx(t *testing.T) {
	wd, err := os.Getwd()
	fmt.Println("wd: ", wd, " err: ", err)
}