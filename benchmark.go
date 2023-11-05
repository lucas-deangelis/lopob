package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
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
	{CommandInput{"ect", []string{"--strict", "-1"}}, "1.png"},
	{CommandInput{"ect", []string{"--strict", "-1"}}, "2.png"},
	{CommandInput{"ect", []string{"--strict", "-1"}}, "3.png"},
	{CommandInput{"ect", []string{"--strict", "-1"}}, "4.png"},
	{CommandInput{"ect", []string{"--strict", "-1"}}, "5.png"},
	{CommandInput{"oxipng", []string{"-o", "0"}}, "1.png"},
	{CommandInput{"oxipng", []string{"-o", "0"}}, "2.png"},
	{CommandInput{"oxipng", []string{"-o", "0"}}, "3.png"},
	{CommandInput{"oxipng", []string{"-o", "0"}}, "4.png"},
	{CommandInput{"oxipng", []string{"-o", "0"}}, "5.png"},
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
	must(os.RemoveAll(workDirectory))
	must(os.MkdirAll(workDirectory, 0755))

	runResults := runAll(runs)
	slices.SortFunc(runResults, func(a, b RunResult) int {
		if a.TargetFilePath == b.TargetFilePath {
			return a.Index - b.Index
		} else {
			return strings.Compare(a.TargetFilePath, b.TargetFilePath)
		}
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	fmt.Fprintln(w, "Tool\tArgs\tImg\tIn\tOut\tSaved\tWall\tSystem\tUser")
	fmt.Fprintln(w, "----\t----\t---\t--\t---\t-----\t----\t------\t----")

	for _, runResult := range runResults {
		fmt.Fprintf(w,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			runResult.CommandToRun.Name,
			strings.Join(runResult.CommandToRun.Args, " "),
			runResult.TargetFilePath,
			Bytes(uint64(runResult.InitialSize)),
			Bytes(uint64(runResult.OptimizedSize)),
			fmt.Sprintf("%.2f%%", (((float64(runResult.InitialSize)-float64(runResult.OptimizedSize))/float64(runResult.InitialSize))*100.0)),
			formatDuration(runResult.WallTime),
			formatDuration(runResult.SystemTime),
			formatDuration(runResult.UserTime),
		)
	}

	w.Flush()
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
