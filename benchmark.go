package main

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

//go:embed test_files
var testFiles embed.FS

const workDirectory = "work"

var images = []string{
	"1.png",
	"2.png",
	"3.png",
	"4.png",
	"5.png",
	"6.png",
	// "7.png",
	// "8.png",
	// "9.png",
	// "10.png",
	// "11.png",
	// "12.png",
	// "13.png",
	// "14.png",
	// "15.png",
	// "16.png",
	// "17.png",
	// "18.png",
	// "19.png",
	// "20.png",
	// "21.png",
	// "22.png",
	// "23.png",
	// "24.png",
	// "25.png",
	// "26.png",
	// "27.png",
}

var commands = []CommandInput{
	{"ect", []string{"--strict", "-1"}},
	// {"ect", []string{"--strict", "-2"}},
	// {"ect", []string{"--strict", "-3"}},
	// {"ect", []string{"--strict", "-4"}},
	{"oxipng", []string{"-o", "0"}},
	// {"oxipng", []string{"-o", "1"}},
	// {"oxipng", []string{"-o", "2"}},
	// {"oxipng", []string{"-o", "3"}},
}

func makeRunInputs(images []string, commands []CommandInput) []RunInput {
	var runInputs []RunInput

	for _, image := range images {
		for _, command := range commands {
			runInputs = append(runInputs, RunInput{
				CommandToRun:   command,
				TargetFilePath: image,
			})
		}
	}

	return runInputs
}

type RunInput struct {
	CommandToRun   CommandInput
	TargetFilePath string
}

type CommandInput struct {
	Name string
	Args []string
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

func (r *RunResult) ToString() []string {
	return []string{
		r.CommandToRun.Name,
		strings.Join(r.CommandToRun.Args, " "),
		r.TargetFilePath,
		Bytes(uint64(r.InitialSize)),
		Bytes(uint64(r.OptimizedSize)),
		fmt.Sprintf("%.2f%%", (((float64(r.InitialSize) - float64(r.OptimizedSize)) / float64(r.InitialSize)) * 100.0)),
		formatDuration(r.WallTime),
		formatDuration(r.SystemTime),
		formatDuration(r.UserTime),
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	must(os.RemoveAll(workDirectory))
	must(os.MkdirAll(workDirectory, 0755))

	runs := makeRunInputs(images, commands)

	runResults := runAllSequential(runs)
	slices.SortFunc(runResults, func(a, b RunResult) int {
		if a.TargetFilePath == b.TargetFilePath {
			return a.Index - b.Index
		} else {
			return strings.Compare(a.TargetFilePath, b.TargetFilePath)
		}
	})

	headers := []string{
		"Tool",
		"Args",
		"Img",
		"In",
		"Out",
		"Saved",
		"Wall",
		"System",
		"User",
	}

	reportFile, err := os.Create("report.csv")
	if err != nil {
		panic(fmt.Sprintf("error creating report.csv, result will only be printed to stdout: %v\n", err))
	}

	out := io.MultiWriter(os.Stdout, reportFile)
	report := csv.NewWriter(out)

	report.Write(headers)
	for _, runResult := range runResults {
		report.Write(runResult.ToString())

	}
	report.Flush()
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(d time.Duration) string {
	totalSeconds := d.Seconds()

	switch {
	case totalSeconds < 1:
		// Use milliseconds for very short durations
		return fmt.Sprintf("%.2fms", d.Seconds()*1000)
	case totalSeconds < 60:
		// Use seconds for durations less than a minute
		return fmt.Sprintf("%.2fs", totalSeconds)
	case totalSeconds < 3600:
		// Use minutes and seconds for durations less than an hour
		minutes := int(totalSeconds / 60)
		seconds := int(totalSeconds) % 60
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	default:
		// Use hours, minutes, and seconds for durations of an hour or more
		hours := int(totalSeconds / 3600)
		minutes := int(totalSeconds/60) % 60
		seconds := int(totalSeconds) % 60
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	}
}

func runAllSequential(runs []RunInput) []RunResult {
	var runResults []RunResult

	for i, run := range runs {
		data, err := runOne(i, run)
		runResults = append(runResults, RunResult{
			Index:    i,
			RunInput: run,
			RunData:  data,
			Err:      err,
		})
	}

	return runResults
}

func runAllConcurrent(runs []RunInput) []RunResult {
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
	if err != nil {
		return data, err
	}

	data.WallTime = time.Since(start)
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
