package filesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func Clean(containerDir string) {
	fmt.Println("CLEANING")
	// os.RemoveAll(containerDir)
}

func FindBinary(command string) (string, error) {
	if filepath.Base(command) != command {
		return command, nil
	}

	pathList := os.Getenv("PATH")
	for _, directory := range filepath.SplitList(pathList) {
		//  it works if it doesn't find the binary in $PATH
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

func CopyFile(source, target string) error {
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

func CreateContainerFileSystem(containerName, command string) (string, string, error) {
	binPath, err := FindBinary(command)
	if err != nil {
		fmt.Printf("Error finding binary: %s\n", err)
		return "", "", err
	}

	containerDir := path.Join(must(os.Getwd()), "tmp", containerName)
	fmt.Println("Create container dir at ", containerDir)
	err = os.MkdirAll(containerDir, 0755)
	if err != nil {
		fmt.Printf("Error creating temp directory for container: %s\n", err)
		return "", "", err
	}

	containerBinPath := path.Join(containerDir, binPath)
	if err := CopyFile(binPath, containerBinPath); err != nil {
		fmt.Printf("Error copying file: %s\n", err)
		return "", "", err
	}

	if err := CopyFile("/dev/null", path.Join(containerDir, "/dev/null")); err != nil {
		fmt.Printf("Error copying file: %s\n", err)
		return "", "", err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		Clean(containerDir)
	}()

	return containerDir, binPath, nil
}

func ParsePort(portsArg string) (portPair [2]uint16, err error) {
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

func must(ret string, err error) string {
	if err != nil {
		os.Exit(1)
	}
	return ret
}
