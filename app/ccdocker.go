package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func main() {

	// Parse args
	if len(os.Args) < 3 || os.Args[1] != "run" {
		log.Fatalf("\nInvalid command: %s\n\nUse: \t ccdocker run \"<container>\" \"<command>\"\n\n", strings.Join(os.Args, " "))
	}

	fmt.Println("Origin PID:: ", os.Getpid())
	runCommand(os.Args[2], os.Args[3])
}

func runCommand(containerName string, command string) {
	fmt.Println("Running command: ", command)

	cmd := exec.Command("/bin/sh", "-c", "hostname "+containerName+";"+command)

	// Set process in new UTS namespace so hostname can be changed
	// without impacting outside environment
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error running command: %s", err)
	}
	fmt.Println("Child PID: ", cmd.Process.Pid)

	//Set signal handlers to forward signals to child process

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Error waiting for command: %s", err)
	}
}
