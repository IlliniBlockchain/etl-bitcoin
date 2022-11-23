package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseCSVMsg(t *testing.T) {
	fileKey := "test"
	msg := NewBaseCSVMsg(fileKey)
	assert.Equal(t, fileKey, msg.FileKey())
}

func TestBaseCSVMsg_Wait(t *testing.T) {
	fileKey := "test"
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "success",
			err:  nil,
		},
		{
			name: "error",
			err:  assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewBaseCSVMsg(fileKey)
			msg.finish(tt.err)
			assert.Equal(t, msg.Wait(), tt.err)
		})
	}
}

func TestNewCSVInsertMsg(t *testing.T) {
	fileKey := "test"
	msg := NewCSVInsertMsg(fileKey, nil)
	assert.Equal(t, fileKey, msg.FileKey())
}
