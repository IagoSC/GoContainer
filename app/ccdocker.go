package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"ccdocker/app/filesystem"
)

func main() {
	fmt.Println(os.Args)

	// Parse args
	if len(os.Args) < 4 {
		log.Fatalf("\nInvalid command: %s\n\nUse: \t ccdocker run \"<container>\" \"<command>\"\n\n", strings.Join(os.Args, " "))
	}

	fmt.Println("Origin PID:: ", os.Getpid())
	runCommand(os.Args[2], os.Args[3], os.Args[4:]...)
}

func runCommand(containerName string, command string, args ...string) {
	fmt.Println("Running command: ", command)

	containerDir, binPath, err := filesystem.CreateContainerFileSystem(containerName, command)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("Container root created")
	fmt.Print("\n\n")
	containerBinPath := path.Join(containerDir, binPath)

	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Set process in new UTS namespace so hostname can be changed without impacting outside environment
		// Set process in new mount namespace so it can have its own filesystem,
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
		// Set process with new root,

		// Leave it as is for a while, less variables to tackle
		Chroot: "tmp/container",
		//TODO TRY SET ENV
	}

	fmt.Println("ContainerDir: ", containerDir)
	fmt.Println("BinPath: ", binPath)
	fmt.Println("Command Path: ", cmd.Path)
	fmt.Println("ContainerBinPath: ", containerBinPath)
	fmt.Println("CMD DIR: ", cmd.Dir)

	fmt.Printf("Running command: %s %v\n", command, args)
	fmt.Println()

	fmt.Println("Run Command")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		filesystem.Clean(containerDir)
		log.Fatalf("Error running command: %s\n", err)
	}

	fmt.Println("Child PID: ", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %s", err)
	}
}
