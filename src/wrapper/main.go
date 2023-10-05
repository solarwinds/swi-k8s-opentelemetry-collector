/*
The goal of this wrapper is to detect panic that indicates corruption of checkpoints folder
(which can happend during kernel crash or system restart) and in case it is detected it removes checkpoint folder
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	panicMessage = "panic: assertion failed: Page expected to be"
)

func main() {
	checkpointDir := os.Getenv("CHECKPOINT_DIR")
	if checkpointDir == "" {
		fmt.Println("Error: CHECKPOINT_DIR environment variable not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Error: Missing command arguments")
		os.Exit(1)
	}

	cmd := exec.Command(os.Args[1], os.Args[2:]...)

	// Setup stderr pipe to capture output
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Goroutine to monitor the output
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)  // Forward the output to stdout
			if strings.Contains(line, panicMessage) {
				fmt.Println("Specific panic detected, deleting all files in checkpoint folder...")
				err := filepath.Walk(checkpointDir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if path != checkpointDir {  // Skip the root directory
						return os.RemoveAll(path)
					}
					return nil
				})
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
		}
	}()

	// Wait for the command to exit
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Wait for the output monitoring goroutine to finish
	wg.Wait()
}