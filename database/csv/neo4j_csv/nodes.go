package neo4j_csv

import (
	"strconv"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

type csvBlockHeader struct {
	*types.BlockHeader
}

func (bh csvBlockHeader) Headers() []string {
	return []string{
		"blockID:id",
		"height:int",
		"time:int",
		"size:int",
		"difficulty:double",
		"nonce:int",
	}
}

func (bh csvBlockHeader) Row() []string {
	return []string{
		bh.Hash(),
		strconv.FormatInt(bh.Height(), 10),
		strconv.FormatInt(bh.Time(), 10),
		strconv.FormatInt(int64(bh.Size()), 10),
		strconv.FormatFloat(bh.Difficulty(), 'f', -1, 64),
		strconv.FormatUint(uint64(bh.Nonce()), 10),
	}
}

type csvTransaction struct {
	*types.Transaction
}

func (tx csvTransaction) Headers() []string {
	return []string{
		"txID:id",
		"size:int",
		"time:int",
		"lockTime:int",
	}
}

func (tx csvTransaction) Row() []string {
	return []string{
		tx.TxID(),
		strconv.FormatInt(int64(tx.Size()), 10),
		strconv.FormatInt(tx.Time(), 10),
		strconv.FormatUint(uint64(tx.LockTime()), 10),
	}
}

type csvVout struct {
	*types.Vout
	tx *types.Transaction
}

func (vout csvVout) Headers() []string {
	return []string{
		"outputID:id",
		"index:int",
		"value:double",
	}
}

func (vout csvVout) Row() []string {
	return []string{
		vout.tx.TxID() + "_" + strconv.FormatInt(int64(vout.N()), 10),
		strconv.FormatInt(int64(vout.N()), 10),
		strconv.FormatFloat(vout.Value(), 'f', -1, 64),
	}
}

type csvCoinbaseVout struct {
	*types.BlockHeader
}

func (bh csvCoinbaseVout) Headers() []string {
	return csvVout{}.Headers()
}

func (bh csvCoinbaseVout) Row() []string {
	return []string{
		bh.Hash() + "_coinbase",
		"0",
		strconv.FormatFloat(bh.Reward(), 'f', -1, 64),
	}
}

type csvAddress struct {
	address string
}

func (addr csvAddress) Headers() []string {
	return []string{
		"addressID:id",
	}
}

func (addr csvAddress) Row() []string {
	return []string{
		addr.address,
	}
}
