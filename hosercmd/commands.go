package hosercmd

// Hoser commands are accepted by the supervisor to control how processes are created and supervised.
// There are also a variety of commands to query info, manipulate pipelines in realtime, etc..
// A hoser cmd file looks like:
//
//  pipeline {"id": "/example"} # creates pipeline
// 	start {"id": "/example/grep0", "exe": "grep", "args": ["-v", "$filter"], "ports": {"filter": {"dir": "in"}}}
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

//easyjson:json
type Start struct {
	Id      string
	ExeFile string `json:"exe"`
	Argv    []string
	Ports   map[string]Port
}

type Dir string

const (
	DirOut = "out"
	DirIn  = "in"
)

type Port struct {
	Dir Dir
}

func (sb *Start) Code() Code {
	return CodeStart
}
