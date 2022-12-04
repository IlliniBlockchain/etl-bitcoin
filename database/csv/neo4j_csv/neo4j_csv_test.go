package neo4j_csv

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	_ database.Database = (*Database)(nil)
	_ database.DBTx     = (*DBTx)(nil)
)

func TestDatabase_LastBlockNumber(t *testing.T) {
	tests := []struct {
		name    string
		opts    database.DBOptions
		want    int64
		wantErr bool
	}{
		{
			name: "success",
			opts: map[string]interface{}{"blocks": "testdata/blocks.csv"},
			want: 2,
		},
		{
			name: "no_records",
			opts: map[string]interface{}{"blocks": "testdata/empty_test.csv"},
			want: 0,
		},
		{
			name:    "no_block_height",
			opts:    map[string]interface{}{"blocks": "testdata/no_block_height.csv"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(context.Background(), tt.opts)
			assert.NoError(t, err)
			got, err := db.LastBlockNumber()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, db.Close())
		})
	}
	assert.NoError(t, cleanTestFiles())
}

type DBTxTestSuite struct {
	suite.Suite
	db   database.Database
	dbTx database.DBTx
}

func (s *DBTxTestSuite) SetupTest() {
	var err error
	opts := make(map[string]interface{})
	for _, fileKey := range fileKeys {
		// Point filepaths to testdata with _test.csv suffix for easy clean up during tear down.
		opts[fileKey] = filepath.Join("testdata", fileKey+"_test.csv")
	}
	s.db, err = NewDatabase(context.Background(), opts)
	s.NoError(err)
	s.dbTx, err = s.db.NewDBTx()
	s.NoError(err)
}

func (s *DBTxTestSuite) TearDownTest() {
	s.NoError(s.db.Close())
	s.NoError(cleanTestFiles())
}

func (s *DBTxTestSuite) TestDBTx_AddBlockHeader() {
	testBlockHeaders := []btcjson.GetBlockVerboseTxResult{
		{
			Hash:       "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
			Height:     0,
			Time:       1231469664,
			Size:       0,
			Difficulty: 0.0,
			Nonce:      0,
		},
		{
			Hash:         "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
			Height:       1,
			Time:         1231469665,
			Size:         0,
			Difficulty:   0.0,
			Nonce:        0,
			PreviousHash: "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		},
		{
			Hash:         "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd",
			Height:       2,
			Time:         1231469666,
			Size:         0,
			Difficulty:   0.0,
			Nonce:        0,
			PreviousHash: "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
		},
	}

	for _, rawBlockHeader := range testBlockHeaders {
		blockHeader := types.NewBlockHeaderFromVerboseTx(rawBlockHeader)
		s.dbTx.AddBlockHeader(blockHeader)
	}
	s.NoError(s.dbTx.Commit())
	s.FilesEqual("testdata/blocks.csv", "testdata/blocks_test.csv")
	s.FilesEqual("testdata/chain.csv", "testdata/chain_test.csv")
	s.FilesEqual("testdata/coinbase.csv", "testdata/coinbase_test.csv")
}

func (s *DBTxTestSuite) TestDBTx_AddTransaction() {
	testTransactions := []btcjson.TxRawResult{
		{
			Txid:      "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
			Size:      134,
			Time:      1231469666,
			LockTime:  0,
			BlockHash: "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd",
			Vin: []btcjson.Vin{
				{
					Coinbase: "04ffff001d010b",
					Sequence: 4294967295,
				},
			},
			Vout: []btcjson.Vout{
				{
					Value: 50,
					N:     0,
					ScriptPubKey: btcjson.ScriptPubKeyResult{
						Asm:  "047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77 OP_CHECKSIG",
						Hex:  "41047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77ac",
						Type: "pubkey",
					},
				},
			},
		},
		{
			Txid:      "cc455ae816e6cdafdb58d54e35d4f46d860047458eacf1c7405dc634631c570d",
			Size:      1963,
			Time:      1231469666,
			LockTime:  0,
			BlockHash: "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd",
			Vin: []btcjson.Vin{
				{
					Txid: "3ead633462a2c980020ffae61d7ccecdc23fda54c022352ea337939da4646c37",
					Vout: 0,
					ScriptSig: &btcjson.ScriptSig{
						Asm: "304402200b78e195f1eb150a52ade3e1e0c593b2534ed3bf4236de4fedb5c8fe7171f3bf02202d63b6c3bd58aa91183a50afb445561854b4bebb6977500f85e61a75b0aa7403[ALL] 02f3ae2c5c5c9616f9e27df9b823af2c748564203afc240a43b8f054dab83c7139",
						Hex: "47304402200b78e195f1eb150a52ade3e1e0c593b2534ed3bf4236de4fedb5c8fe7171f3bf02202d63b6c3bd58aa91183a50afb445561854b4bebb6977500f85e61a75b0aa7403012102f3ae2c5c5c9616f9e27df9b823af2c748564203afc240a43b8f054dab83c7139",
					},
					Sequence: 4294967295,
				},
			},
			Vout: []btcjson.Vout{
				{
					Value: 0.0001,
					N:     0,
					ScriptPubKey: btcjson.ScriptPubKeyResult{
						Asm:       "OP_DUP OP_HASH160 8d1ec2350813b2a071353e16b41e884647405d3d OP_EQUALVERIFY OP_CHECKSIG",
						Hex:       "76a9148d1ec2350813b2a071353e16b41e884647405d3d88ac",
						Addresses: []string{"1DsB8CHc7fa87GUm6qs6YGoCRgQpBo2TJZ"},
						Type:      "pubkeyhash",
					},
				},
			},
		},
	}
	for _, rawTx := range testTransactions {
		tx := types.NewTransaction(rawTx)
		s.dbTx.AddTransaction(tx)
	}
	s.NoError(s.dbTx.Commit())
	s.FilesEqual("testdata/transactions.csv", "testdata/transactions_test.csv")
	s.FilesEqual("testdata/outputs.csv", "testdata/outputs_test.csv")
	s.FilesEqual("testdata/addresses.csv", "testdata/addresses_test.csv")
	s.FilesEqual("testdata/in.csv", "testdata/in_test.csv")
	s.FilesEqual("testdata/out.csv", "testdata/out_test.csv")
	s.FilesEqual("testdata/include.csv", "testdata/include_test.csv")
	s.FilesEqual("testdata/locked.csv", "testdata/locked_test.csv")

}

func TestDBTxTestSuite(t *testing.T) {
	suite.Run(t, new(DBTxTestSuite))
}

func (s *DBTxTestSuite) FilesEqual(wantFile, gotFile string) {
	wantFileBytes, err := os.ReadFile(wantFile)
	s.NoError(err)
	gotFileBytes, err := os.ReadFile(gotFile)
	s.NoError(err)
	s.Equal(string(wantFileBytes), string(gotFileBytes))
}

func cleanTestFiles() error {
	files, err := os.ReadDir("testdata")
	if err != nil {
		return err
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), "_test.csv") {
			continue
		}
		if err := os.Remove(filepath.Join("testdata", file.Name())); err != nil {
			return err
		}
	}
	return nil
}
