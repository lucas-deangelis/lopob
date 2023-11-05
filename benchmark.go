package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

//go:embed test_files
var testFiles embed.FS

const workDirectory = "work"

type RunInput struct {
	CommandToRun   CommandInput
	TargetFilePath string
}

type CommandInput struct {
	Name string
	Args []string
}

var runs = []RunInput{
	{CommandInput{"ect", []string{"-strict", "-3"}}, "1.png"},
	{CommandInput{"ect", []string{"-strict", "-3"}}, "2.png"},
	{CommandInput{"ect", []string{"-strict", "-3"}}, "3.png"},
	{CommandInput{"oxipng", []string{"-o", "2"}}, "1.png"},
	{CommandInput{"oxipng", []string{"-o", "2"}}, "2.png"},
	{CommandInput{"oxipng", []string{"-o", "2"}}, "3.png"},
}

func main() {
	runAll(runs)
}

func runAll(runs []RunInput) {
	var wg sync.WaitGroup

	for i, run := range runs {
		wg.Add(1)
		go func(runIndex int, run RunInput) {
			defer wg.Done()
			_, err := runOne(i, run)
			if err != nil {
				panic(err)
			}
		}(i, run)
	}

	wg.Wait()
}

func runOne(runIndex int, run RunInput) (string, error) {
	src, err := testFiles.ReadFile("test_files/" + run.TargetFilePath)
	if err != nil {
		return "", err
	}

	tempFilePath := fmt.Sprintf("%s/%d.png", workDirectory, runIndex)

	err = os.WriteFile(tempFilePath, src, 0644)
	if err != nil {
		return "", err
	}

	nameAndArgs := append([]string{run.CommandToRun.Name}, run.CommandToRun.Args...)
	nameAndArgsAndTargetFile := append(nameAndArgs, tempFilePath)

	cmd := exec.Command("time", nameAndArgsAndTargetFile...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return "", nil
}
