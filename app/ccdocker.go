package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
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

	switch os.Args[1] {
	case "run":
		fmt.Println("Origin PID:: ", os.Getpid())
		runCommand(os.Args[2], os.Args[3])
	case "child":
		fmt.Println("Child PID:: ", os.Getpid())
		// runChild(os.Args[2], os.Args[3])
	}
}

// This is just for testing
// func runChild(containerName, command string) {
// 	currentDir, _ := os.Getwd()
// 	fmt.Println(currentDir)

// 	syscall.Chroot()
// }

func runCommand(containerName string, command string) {
	fmt.Println("Running command: ", command)

	containerDir, binPath, _ := filesystem.CreateContainerFileSystem(containerName, command)
	// A try to fix
	// still doesn't clean
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		filesystem.Clean(containerDir)
	}()

	fmt.Println("Container root created")
	fmt.Print("\n\n")

	currentDir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	containerBinPath := path.Join(currentDir, containerDir, binPath)

	cmd := exec.Command(command, "")

	fmt.Println("ContainerDir: ", containerDir)
	fmt.Println("BinPath: ", binPath)
	fmt.Println("Command Path: ", cmd.Path)
	fmt.Println("ContainerBinPath: ", containerBinPath)

	fmt.Printf("Running command: %s\n", command)
	fmt.Println()

	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Set process in new UTS namespace so hostname can be changed without impacting outside environment
		// Set process in new mount namespace so it can have its own filesystem,
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS,
		// Set process with new root,
		// Chroot: containerDir, // This is not used with /proc/self/exe
	}

	fmt.Println("Run Command")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error running command: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Child PID: ", cmd.Process.Pid)

	//Set signal handlers to forward signals to child process
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %s", err)
	}
}
