package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// hoser-xargs takes in a stream of arguments and creates a process for each line
// passing that line as an argument. The stdout and stderr of each process is written
// as filenames to stdout. If any errors occur, they are written to stderr.

const MaxLineSize = 5024

var (
	replacementToken = flag.String("tok", "{}", "Replacement token (token will be replaced with line in stdin)")
	errs             = flag.String("errs", "", "A destination file to write error streams that child processes produce")
	parallelism      = flag.Int("p", 1, "Number of concurrent processes to execute. If zero, maximum number of cores (default 1)")
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)

	// cmdArgs := flag.Args()
	// var errsFd *os.File
	// if *errs != "" {
	// 	errsFd, err = os.Open(*errs)
	// 	if err != nil {
	// 		log.Printf("bad errs flag '%s': %v", *errs, err)
	// 	}
	// }

	buf := bufio.NewReaderSize(os.Stdin, MaxLineSize)
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			log.Printf("stdin closed: %v", err)
			return
		}
		line = strings.TrimSuffix(line, "\n")
		if len(line) == 0 {
			continue // skip empty lines
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}
}

type job struct {
	GivenArgs []string
	Line      string
	LineNum   int
}

func execJob(j job, outCh, errCh chan string) {
	cmd := execLine(j.GivenArgs, j.Line)
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		log.Printf("[line %d]: could not create pipe for stdout", j.LineNum)
		return
	}
	defer stdoutR.Close()
	defer stdoutW.Close()

	cmd.Stdout = stdoutW
	var stderrR, stderrW *os.File
	if errCh != nil {
		stderrR, stderrW, err = os.Pipe()
		if err != nil {
			log.Printf("[line %d]: could not create pipe for stderr", j.LineNum)
			return
		}
		cmd.Stderr = stderrW
		errCh <- stderrR.Name()
	}
	outCh <- stdoutR.Name()

	err = cmd.Run()
	if err != nil {
		log.Printf("[line %d] fail: pid '%v' run: %v", j.LineNum, cmd.Process.Pid, err)
		return
	}
}

func execLine(cmdArgs []string, replacement string) *exec.Cmd {
	for i := range cmdArgs {
		cmdArgs[i] = strings.ReplaceAll(cmdArgs[i], *replacementToken, replacement)
	}
	return exec.Command(cmdArgs[0], cmdArgs[1:]...)
}
