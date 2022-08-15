package hosercmd

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

var (
	ErrTooLong  = errors.New("id too long (too many slashes)")
	ErrTooShort = errors.New("id must have pipeline")
)

// Ident is an identifier in hosercmd. An identifier can have three different scopes:
// pipeline: /pipelineID
// process/var: /pipelineID/varID OR /pipelineID/processID
// port: /pipelineID/processID[portID]
type Ident struct {
	Pipeline, Node, Port string // any of these values can be empty
}

func (i Ident) String() string {
	var sb strings.Builder
	sb.WriteByte('/')
	sb.WriteString(i.Pipeline)
	if i.Node != "" || i.Port != "" {
		sb.WriteByte('/')
		sb.WriteString(i.Node)
		if i.Port != "" {
			sb.WriteByte('[')
			sb.WriteString(i.Port)
			sb.WriteByte(']')
		}
	}
	return sb.String()
}

func ParseId(id string) (Ident, error) {
	if id[0] != '/' {
		return Ident{}, fmt.Errorf("id must start with /")
	}

	parts, err := split(id)
	if err != nil {
		return Ident{}, fmt.Errorf("bad id: %w", err)
	}

	l := len(parts)
	switch {
	case l == 1:
		return Ident{Pipeline: parts[0]}, nil
	case l == 2:
		return Ident{Pipeline: parts[0], Node: parts[1]}, nil
	case l == 3:
		return Ident{Pipeline: parts[0], Node: parts[1], Port: parts[2]}, nil
	case l > 3:
		return Ident{}, ErrTooLong
	default:
		return Ident{}, ErrTooShort
	}
}

func split(p string) (parts []string, err error) {
	if p[0] != '/' {
		panic("path must start with /")
	}

	for p != "/" {
		dir, file := path.Split(p)
		p = path.Dir(dir)
		parts = append(parts, file)
	}
	parts = reverse(parts)

	if len(parts) > 2 {
		return nil, ErrTooLong
	}

	if len(parts) == 2 {
		process, port := parsePort(parts[1])
		parts[1] = process
		if port != "" {
			parts = append(parts, port)
		}
	}
	return parts, nil
}

func parsePort(combined string) (process, port string) {
	j := strings.LastIndexByte(combined, '[')
	if j < 0 {
		return combined, ""
	}

	k := strings.LastIndexByte(combined, ']')
	if k < 0 || k < j || k != len(combined)-1 {
		return combined, ""
	}
	return combined[:j], combined[j+1 : k]
}

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
