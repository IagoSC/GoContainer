package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Running ./test AAAAAAAAAA")
	cmd := exec.Command("sh")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR starting: ", err)
		os.Exit(1)
	}
	// time.Sleep(10 * time.Minute)

	if err := cmd.Wait(); err != nil {
		fmt.Println("ERROR waiting: ", err)
		os.Exit(1)
	}
}
