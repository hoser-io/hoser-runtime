package hosercmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var reallyLongString = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`

func TestReadFiles(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    []Command
		wantErr bool
	}{
		{
			"start", `
	start {"id": "1"}	
// comment here
start {"id": "2"}
start {"id": "3"}`,
			[]Command{
				{Start, &StartBody{Id: "1"}},
				{Start, &StartBody{Id: "2"}},
				{Start, &StartBody{Id: "3"}},
			},
			false,
		},
		{
			"bad line", `
start {"id": "1"}	
notacode`,
			[]Command{{Start, &StartBody{Id: "1"}}},
			true,
		},
		{
			"bad comment", `
start {"id": "1"}	
/`,
			[]Command{{Start, &StartBody{Id: "1"}}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.args)
			cmds, err := ReadFiles(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, cmds)
		})
	}

}

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    Command
		wantErr bool
	}{
		{"start", `start {"id":"a"}`, Command{Code: Start, Body: &StartBody{Id: "a"}}, false},
		{"bad code", `thisisbad {"id":"a"}`, Command{}, true},
		{"no body", `start`, Command{}, true},
		{"body not json", `start {`, Command{}, true},
		{"too long", reallyLongString, Command{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read([]byte(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
