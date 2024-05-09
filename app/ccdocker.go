package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	// Parse args
	if len(os.Args) < 3 || os.Args[1] != "run" {
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

func runCommand(containerName string, command string) {
	fmt.Println("Running command: ", command)

	// TODO Make this using "os" package
	binPath, err := findBinary(command)
	if err != nil {
		log.Fatalf("Error finding binary: %s", err)
	}
	containerDir := "/tmp/" + containerName + "/"
	if err := copyFile(binPath, containerDir+filepath.Base(binPath)); err != nil {
		log.Fatalf("Error copying file: %s", err)
	}
	// NOT CALLED WHEN SIGTERM IS SENT
	defer os.RemoveAll(containerDir)

	cmd := exec.Command("/bin/sh", "-c", "hostname "+containerName+";"+command)

	// Set process in new UTS namespace so hostname can be changed without impacting outside environment
	// Set process in new mount namespace so it can have its own filesystem,
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS,
	// }

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error running command: %s", err)
	}
	fmt.Println("Child PID: ", cmd.Process.Pid)
	fmt.Println()

	//Set signal handlers to forward signals to child process

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Error waiting for command: %s", err)
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
