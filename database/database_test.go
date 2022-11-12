package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOpt(t *testing.T) {
	type args struct {
		opts DBOptions
		key  string
		def  int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "return option",
			args: args{
				opts: DBOptions{"int": 1},
				key:  "int",
				def:  0,
			},
			want: 1,
		},
		{
			name: "return default",
			args: args{
				opts: DBOptions{"not_int": "1"},
				key:  "int",
				def:  0,
			},
			want: 0,
		},
		{
			name: "wrong type",
			args: args{
				opts: DBOptions{"int": "1"},
				key:  "int",
				def:  0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOpt(tt.args.opts, tt.args.key, tt.args.def)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
