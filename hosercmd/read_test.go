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
				&Start{Id: "1"},
				&Start{Id: "2"},
				&Start{Id: "3"},
			},
			false,
		},
		{
			"bad line", `
start {"id": "1"}	
notacode`,
			[]Command{&Start{Id: "1"}},
			true,
		},
		{
			"bad comment", `
start {"id": "1"}	
/`,
			[]Command{&Start{Id: "1"}},
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
		{"start", `start {"id":"a"}`, &Start{Id: "a"}, false},
		{"bad code", `thisisbad {"id":"a"}`, nil, true},
		{"no body", `start`, nil, true},
		{"body not json", `start {`, &Start{}, true},
		{"too long", reallyLongString, nil, true},
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
