package hosercmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmd_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantProc StartBody
		wantErr  bool
	}{
		{"no args", `{"id":"/pipeline/a","exe":"awk","argv":[]}`, StartBody{Id: "/pipeline/a", ExeFile: "awk"}, false},
		{"string args", `{"argv":["a","b","c"]}`, StartBody{Argv: []Arg{StringArg("a"), StringArg("b"), StringArg("c")}}, false},
		{"in/out args", `{"argv":[{"out":"a"},"c",{"in":"b"}]}`, StartBody{Argv: []Arg{&NamedArg{Out: "a"}, StringArg("c"), &NamedArg{In: "b"}}}, false},
		{"unknown field", `{"args":[]}`, StartBody{}, true},
		{"bad json", `{"argv":[{"out"}]}`, StartBody{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var original StartBody
			var err error
			if err = original.UnmarshalJSON([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Errorf("ProcCmd.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			assert.Equal(t, tt.wantProc, original)

			marshalled, err := original.MarshalJSON()
			assert.NoError(t, err)
			var after StartBody
			err = after.UnmarshalJSON(marshalled)
			assert.NoError(t, err)
			assert.Equal(t, original, after)
		})
	}
}

func TestParseArgv(t *testing.T) {
	args, err := ParseArgv([]byte(`["a", "b", {"in": "test"}]`))
	assert.Len(t, args, 3)
	assert.NoError(t, err)
}

func Test_ReadsCommands(t *testing.T) {
	tests := []struct {
		wantCode Code
		line     string
	}{
		{Start, `start {"id":"/pipeline/a","exe":"awk","argv":[]}`},
		{Pipeline, `pipeline {"id":"/pipeline"}`},
		{Pipe, `pipe {"src":"/pipeline/v1","dst":"/pipeline/v2"}`},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.line), func(t *testing.T) {
			cmd, err := Read([]byte(tt.line))
			assert.Equal(t, tt.wantCode, cmd.Code)
			assert.NoError(t, err)

			bytes, err := cmd.Body.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.line, fmt.Sprintf("%s %s", tt.wantCode, bytes))
		})
	}
}
