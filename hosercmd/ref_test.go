package hosercmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		args    args
		want    Ident
		wantErr bool
	}{
		{args{"/test"}, Ident{Pipeline: "test"}, false},
		{args{"/test/process"}, Ident{Pipeline: "test", Node: "process"}, false},
		{args{"/test/process[port]"}, Ident{Pipeline: "test", Node: "process", Port: "port"}, false},
		{args{"/test/"}, Ident{Pipeline: "test"}, false},
		{args{"/"}, Ident{}, true},
		{args{"whatisthis"}, Ident{}, true},
		{args{"/bad/por[]t"}, Ident{Pipeline: "bad", Node: "por[]t"}, false},
		{args{"/too/many/paths"}, Ident{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.args.id, func(t *testing.T) {
			got, err := ParseId(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseId(%s) error = %v, wantErr %v", tt.args.id, err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIdString(t *testing.T) {
	assert.Equal(t, "/test/process[port]", Ident{"test", "process", "port"}.String())
}
