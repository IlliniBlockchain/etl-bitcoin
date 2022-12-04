package loader

import (
	"encoding/json"
	"os"
	"testing"
)

// Requires integration test or dummy client
func TestLoaderManager(t *testing.T) {

}

func TestBlockRangeHandler(t *testing.T) {
	// type args struct {
	// 	client client.Client
	// 	msg    *LoaderMsg[BlockRange]
	// }
	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	want    *LoaderMsg[[]*chainhash.Hash]
	// 	wantErr bool
	// }{
	// 	{
	// 		name: "return option",
	// 		args: args{
	// 			client: nil,
	// 			msg: nil,
	// 		},
	// 		want: 1,
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err :=
	// 		if tt.wantErr {
	// 			assert.Error(t, err)
	// 			return
	// 		}
	// 		assert.Equal(t, tt.want, got)
	// 	})
	// }

}

func TestBlockHashHandler(t *testing.T) {

}

func TestBlockHandler(t *testing.T) {

}

func parseTestData[T any](filename string, v *T) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}
