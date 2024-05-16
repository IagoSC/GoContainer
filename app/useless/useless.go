package useless

import (
	"fmt"
	"os/exec"
	"reflect"
)

func DeepCompare(cmd1, cmd2 *exec.Cmd) {
	fmt.Println("COMPARING")
	v1 := reflect.ValueOf(*cmd1)
	v2 := reflect.ValueOf(*cmd2)

	typeOfS1 := v1.Type()
	for i := 0; i < v1.NumField(); i++ {
		fieldName := typeOfS1.Field(i).Name
		fmt.Println("Field Name: ", fieldName)
		value1 := v1.Field(i).Interface()
		value2 := v2.Field(i).Interface()

		if reflect.DeepEqual(value1, value2) {
			fmt.Println("DIFF:: ")
			fmt.Println("Value 1: ", value1)
			fmt.Println("Value 2: ", value2)
		} else {
			fmt.Println("THEY ARE EQUAL")
		}
		fmt.Println()
	}
}
