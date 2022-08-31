package hosercmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	MaxCodeLen = 128 // 128 chars is a conservative upper limit on command code len (e.g. start)
)

func ReadFiles(r io.Reader) (cmds []Command, err error) {
	s := bufio.NewScanner(r)
	lineNo := 0
	for s.Scan() {
		lineNo += 1
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			continue // skip whitespace only lines
		}
		if len(line) >= 2 && bytes.Compare(line[:2], []byte("//")) == 0 {
			continue // comment
		}
		cmd, err := Read(line)
		if err != nil {
			return cmds, &Error{LineNumber: lineNo, Context: line, Err: err}
		}
		cmds = append(cmds, cmd)
	}
	return cmds, s.Err()
}

type Error struct {
	LineNumber int
	Context    []byte
	Err        error
}

func (e *Error) Error() string {
	return fmt.Sprintf("syntax error (line %d): %v", e.LineNumber, e.Err)
}

func Read(line []byte) (Command, error) {
	code, rest := readCode(line)
	var cmd Command
	switch code {
	case CodeStart:
		cmd = &Start{}
	case CodePipeline:
		cmd = &Pipeline{}
	case CodeSet:
		cmd = &Set{}
	case CodePipe:
		cmd = &Pipe{}
	case CodeExit:
		cmd = &Exit{}
	default:
		return nil, fmt.Errorf("unrecognized command: %s", code)
	}

	err := cmd.UnmarshalJSON(rest)
	if err != nil {
		return cmd, err
	}
	return cmd, nil
}

func readCode(line []byte) (code Code, rest []byte) {
	for i, v := range line {
		if i > MaxCodeLen {
			return
		}
		isspace := v == '\t' || v == ' '
		if isspace {
			code = Code(line[:i])
		}
		if code != "" && !isspace {
			return code, line[i:]
		}
	}
	return
}
