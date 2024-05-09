package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	fmt.Println(os.Args)

	// Parse args
	if len(os.Args) < 4 || os.Args[1] != "run" {
		// Doesn't work with go run (?)
		log.Fatalf("\nInvalid command: %s\n\nUse: \t ccdocker run \"<container>\" \"<command>\"\n\n", strings.Join(os.Args, " "))
	}

	fmt.Println("Origin PID:: ", os.Getpid())
	runCommand(os.Args[2], os.Args[3])
}

func findBinary(command string) (string, error) {
	pathList := os.Getenv("PATH")
	for _, directory := range filepath.SplitList(pathList) {
		path := filepath.Join(directory, command)
		fileInfo, err := os.Stat(path)
		if err == nil {
			mode := fileInfo.Mode()
			if mode.IsRegular() && mode&0111 != 0 {
				fmt.Println("Path: ", path)
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("binary not found: %s", command)
}

func copyFile(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceInfoFile, err := os.Stat(source)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	if err := targetFile.Chmod(sourceInfoFile.Mode()); err != nil {
		return err
	}
	return nil
}

func createContainerFileSystem(containerName, command string) string {
	binPath, err := findBinary(command)
	if err != nil {
		fmt.Printf("Error finding binary: %s\n", err)
	}

	containerDir, err := os.MkdirTemp("", containerName+"_*")
	if err != nil {
		fmt.Printf("Error creating temp directory for container: %s\n", err)
		os.Exit(1)
	}
	containerBinPath := path.Join(containerDir, binPath)
	if err := copyFile(binPath, containerBinPath); err != nil {
		fmt.Printf("Error copying file: %s\n", err)
		os.Exit(1)
	}
	if err := copyFile("/dev/null", path.Join(containerDir, "/dev/null")); err != nil {
		fmt.Printf("Error copying file: %s\n", err)
		os.Exit(1)
	}

	if err := syscall.Chroot(containerDir); err != nil {
		fmt.Printf("Couldn't run chroot syscall: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Container root created")
	return containerBinPath
}

func runCommand(containerName string, command string) {
	fmt.Println("Running command: ", command)

	containerBinPath := createContainerFileSystem(containerName, command)
	fmt.Println(containerBinPath)
	cmd := exec.Command(containerBinPath)

	// NOT CALLED WHEN CLOSED BY SIGNAL
	// defer os.RemoveAll(containerDir)
	// A try to fix
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		// os.RemoveAll(containerDir)
	}()

	// Set process in new UTS namespace so hostname can be changed without impacting outside environment
	// Set process in new mount namespace so it can have its own filesystem,
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS,
	// }

	fmt.Println("Run Command")

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error running command: %s\n", err)
	}
	fmt.Println("Child PID: ", cmd.Process.Pid)

	//Set signal handlers to forward signals to child process
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %s", err)
	}
}

// type TransformWriter struct {
// 	w io.Writer
// }

// func (tw *TransformWriter) Write(p []byte) (n int, err error) {
// 	transformedOutput := string(p)

// 	// defer tw.w.Write([]byte("\n>> "))
// 	return tw.w.Write([]byte(transformedOutput))
// }

func parsePort(portsArg string) (portPair [2]uint16, err error) {
	ports := strings.Split(portsArg, ":")
	if len(ports) != 2 {
		log.Fatal("Invalid port format, use <host_port>:<container_port>")
	}
	for i, port := range ports {
		portValue, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Invalid port: %s", port)
		}
		portPair[i] = uint16(portValue)
	}
	return
}
