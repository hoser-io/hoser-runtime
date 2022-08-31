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
		wantProc Start
		wantErr  bool
	}{
		{"no args", `{"id":"/pipeline/a","exe":"awk","argv":[]}`, Start{Id: "/pipeline/a", ExeFile: "awk"}, false},
		{"string args", `{"argv":["a","b","c"]}`, Start{Argv: []Arg{StringArg("a"), StringArg("b"), StringArg("c")}}, false},
		{"in/out args", `{"argv":[{"out":"a"},"c",{"in":"b"}]}`, Start{Argv: []Arg{&NamedArg{Out: "a"}, StringArg("c"), &NamedArg{In: "b"}}}, false},
		{"unknown field", `{"args":[]}`, Start{}, true},
		{"bad json", `{"argv":[{"out"}]}`, Start{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var original Start
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
			var after Start
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
		{CodeStart, `start {"id":"/pipeline/a","exe":"awk","argv":[]}`},
		{CodePipeline, `pipeline {"id":"/pipeline"}`},
		{CodePipe, `pipe {"src":"/pipeline/v1","dst":"/pipeline/v2"}`},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.line), func(t *testing.T) {
			cmd, err := Read([]byte(tt.line))
			assert.Equal(t, tt.wantCode, cmd.Code)
			assert.NoError(t, err)

			bytes, err := cmd.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.line, fmt.Sprintf("%s %s", tt.wantCode, bytes))
		})
	}
}
