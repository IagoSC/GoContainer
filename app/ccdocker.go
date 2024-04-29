package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {

	// Parse args
	if len(os.Args) != 3 || os.Args[1] != "run" {
		log.Fatalf("\nInvalid command: %s\n\nUse: \t ccdocker run \"<command>\"\n\n", strings.Join(os.Args, " "))
	}

	runCommand(os.Args[2])

}

type TransformWriter struct {
	w io.Writer
}

func (tw *TransformWriter) Write(p []byte) (n int, err error) {
	transformedOutput := string(p)

	defer tw.w.Write([]byte("\n>> "))
	return tw.w.Write([]byte(transformedOutput))
}

func runCommand(command string) {
	cmdArgs := strings.Split(command, " ")
	fmt.Println("Running command: ", command)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = &TransformWriter{os.Stdout}
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error running command: %s", err)
	}
}

