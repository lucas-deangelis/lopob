package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"sync"
	"time"
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

type RunData struct {
	InitialSize   int64
	OptimizedSize int64
	WallTime      time.Duration
	SystemTime    time.Duration
	UserTime      time.Duration
}

type RunResult struct {
	Index int
	RunInput
	RunData
	Err error
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	must(os.MkdirAll(workDirectory, 0755))
	runResults := runAll(runs)
	slices.SortFunc(runResults, func(a, b RunResult) int {
		return a.Index - b.Index
	})

	for _, runResult := range runResults {
		fmt.Printf("Run %d: %s\n", runResult.Index, runResult.Err)
		fmt.Printf("Initial size: %d bytes\n", runResult.InitialSize)
		fmt.Printf("Optimized size: %d bytes\n", runResult.OptimizedSize)
		fmt.Printf("Wall time: %s\n", runResult.WallTime)
		fmt.Printf("System time: %s\n", runResult.SystemTime)
		fmt.Printf("User time: %s\n", runResult.UserTime)
		fmt.Printf("\n")
	}
}

func runAll(runs []RunInput) []RunResult {
	var wg sync.WaitGroup
	results := make(chan RunResult, len(runs))

	for i, run := range runs {
		wg.Add(1)
		go func(runIndex int, run RunInput) {
			defer wg.Done()
			data, err := runOne(runIndex, run)
			results <- RunResult{
				Index:    runIndex,
				RunInput: run,
				RunData:  data,
				Err:      err,
			}
		}(i, run)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var runResults []RunResult
	for result := range results {
		if result.Err != nil {
			panic(result.Err)
		}
		runResults = append(runResults, result)
	}

	return runResults
}

func runOne(runIndex int, run RunInput) (RunData, error) {
	data := RunData{}

	initialInfos, err := fs.Stat(testFiles, "test_files/"+run.TargetFilePath)
	if err != nil {
		return data, err
	}
	data.InitialSize = initialInfos.Size()
	src, err := testFiles.ReadFile("test_files/" + run.TargetFilePath)
	if err != nil {
		return data, err
	}

	tempFilePath := fmt.Sprintf("%s/%d.png", workDirectory, runIndex)

	err = os.WriteFile(tempFilePath, src, 0644)
	if err != nil {
		return data, err
	}
	defer os.Remove(tempFilePath)

	argsWithTargetPath := append(run.CommandToRun.Args, tempFilePath)

	cmd := exec.Command(run.CommandToRun.Name, argsWithTargetPath...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err = cmd.Run()
	data.WallTime = time.Since(start)

	if err != nil {
		return data, err
	}

	if cmd.ProcessState != nil {
		data.SystemTime = cmd.ProcessState.SystemTime()
		data.UserTime = cmd.ProcessState.UserTime()
	}

	// Get the size of the file after the command ran
	optimizedInfos, err := os.Stat(tempFilePath)
	if err != nil {
		return data, err
	}

	data.OptimizedSize = optimizedInfos.Size()

	return data, nil
}
