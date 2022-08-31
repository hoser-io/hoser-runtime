package hosercmd

import (
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

// Hoser commands are accepted by the supervisor to control how processes are created and supervised.
// There are also a variety of commands to query info, manipulate pipelines in realtime, etc..
// A hoser cmd file looks like:
//
//  pipeline {"id": "/example"} # creates pipeline
// 	start {"id": "/example/grep0", "exe": "grep", "args": ["-v", {"port": "filter"}]}
// 	set {"id": "/example/keyword", "text": "cats"}
// 	set {"id": "/example/stdin", "read": "file://cats.txt"}
// 	set {"id": "/example/stdout", "write": "file://output.txt"}
//  pipe {"id": "/example/reserved"}
// 	pipe {"src": "stdin", "dst": "grep0[stdin]"}
// 	pipe {"src": "/example/grep0[stdout]", "dst": "example/stdout"}
// 	pipe {"src": "/example/keyword", "dst": "/example/grep0[filter]"}
//
// which is just a word (e.g. start) followed by a JSON body describing the arguments.

// Generate marshaling functions for simpler command structs
//go:generate easyjson -snake_case -disallow_unknown_fields commands.go

type Code string

var (
	CodeStart    Code = "start"
	CodePipeline Code = "pipeline"
	CodeSet      Code = "set"
	CodePipe     Code = "pipe"
	CodeExit     Code = "exit"
)

type Command interface {
	UnmarshalJSON(body []byte) error
	MarshalJSON() ([]byte, error)
	Code() Code
}
type Result = Command

//easyjson:json
type Pipeline struct {
	Id string
}

func (b *Pipeline) Code() Code {
	return CodePipeline
}

//easyjson:json
type Set struct {
	Id          string
	Read, Write string // URLs to read and write data to. Read creates a source, Write creates a sink.
	Text        string // A fixed value for sources
}

func (sb *Set) Code() Code {
	return CodeSet
}

func (sb *Set) IsSink() bool {
	return sb.Write != ""
}

func (sb *Set) IsSpout() bool {
	return sb.Text != "" || sb.Read != ""
}

//easyjson:json
type Pipe struct {
	Src, Dst string
}

func (b *Pipe) Code() Code {
	return CodePipe
}

//easyjson:json
type Exit struct {
	When string
}

func (b *Exit) Code() Code {
	return CodeExit
}

type Start struct {
	Id      string
	ExeFile string
	Argv    []Arg
}

func (sb *Start) Code() Code {
	return CodeStart
}

type Arg interface {
	arg()
}

type StringArg string
type NamedArg struct {
	In  string
	Out string
}

func (na *NamedArg) Argname() string {
	if na.In != "" {
		return na.In
	}
	return na.Out
}

func (na *NamedArg) IsIngress() bool {
	return na.In != ""
}

func (sa StringArg) arg() {}
func (na *NamedArg) arg() {}

// Custom marshalling and unmarshalling for ProcCmd because of argv having irregular array args
func (pc *Start) UnmarshalJSON(body []byte) error {
	r := &jlexer.Lexer{Data: body}
	r.Delim('{')
	for !r.IsDelim('}') {
		key := r.String()
		switch key {
		case "id":
			r.WantColon()
			pc.Id = r.String()
			r.WantComma()
		case "exe":
			r.WantColon()
			pc.ExeFile = r.String()
			r.WantComma()
		case "argv":
			r.WantColon()
			pc.unmarshalArgv(r)
			r.WantComma()
		default:
			r.AddError(&jlexer.LexerError{
				Offset: r.GetPos(),
				Reason: "unknown field",
				Data:   key,
			})
		}
	}
	r.Delim('}')
	return r.Error()
}

func ParseArgv(argv []byte) ([]Arg, error) {
	r := &jlexer.Lexer{Data: argv}
	var body Start
	body.unmarshalArgv(r)
	return body.Argv, r.Error()
}

func (pc *Start) unmarshalArgv(r *jlexer.Lexer) {
	r.Delim('[')
	for !r.IsDelim(']') {
		if r.IsDelim('{') {
			r.Delim('{')
			key := r.String()
			r.WantColon()
			if key == "in" {
				pc.Argv = append(pc.Argv, &NamedArg{In: r.String()})
			} else if key == "out" {
				pc.Argv = append(pc.Argv, &NamedArg{Out: r.String()})
			}
			r.WantComma()
			r.Delim('}')
		} else {
			pc.Argv = append(pc.Argv, StringArg(r.String()))
		}
		r.WantComma()
	}
	r.Delim(']')
}

func (pc Start) MarshalJSON() ([]byte, error) {
	var wr jwriter.Writer
	wr.RawByte('{')
	wr.RawString(`"id":`)
	wr.String(pc.Id)
	wr.RawString(`,"exe":`)
	wr.String(pc.ExeFile)
	wr.RawString(`,"argv":`)

	wr.RawByte('[')
	for i, arg := range pc.Argv {
		switch v := arg.(type) {
		case StringArg:
			wr.String(string(v))
		case *NamedArg:
			if v.In != "" {
				wr.RawString(`{"in":`)
				wr.String(v.In)
			} else if v.Out != "" {
				wr.RawString(`{"out":`)
				wr.String(v.Out)
			} else {
				wr.RawByte('{')
			}
			wr.RawByte('}')
		}
		if i != len(pc.Argv)-1 {
			wr.RawByte(',')
		}
	}
	wr.RawString("]}")
	return wr.BuildBytes()
}
