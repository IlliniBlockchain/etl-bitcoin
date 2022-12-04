package neo4j_csv

import (
	"strconv"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

type csvChainRelation struct {
	*types.BlockHeader
}

func (bh csvChainRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (bh csvChainRelation) Row() []string {
	return []string{
		bh.Hash(),
		bh.PreviousHash(),
	}
}

type csvCoinbaseRelation struct {
	*types.BlockHeader
}

func (bh csvCoinbaseRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (bh csvCoinbaseRelation) Row() []string {
	return []string{
		bh.Hash(),
		bh.Hash() + "_coinbase",
	}
}

type csvIncludeRelation struct {
	*types.Transaction
}

func (tx csvIncludeRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (tx csvIncludeRelation) Row() []string {
	return []string{
		tx.TxID(),
		tx.BlockHash(),
	}
}

type csvInRelation struct {
	*types.Vin
	tx *types.Transaction
}

func (in csvInRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (in csvInRelation) Row() []string {
	return []string{
		in.TxID() + strconv.FormatInt(int64(in.Vout()), 10),
		in.tx.TxID(),
	}
}

type csvCoinbaseInRelation struct {
	*types.Transaction
}

func (tx csvCoinbaseInRelation) Headers() []string {
	return csvInRelation{}.Headers()
}

func (tx csvCoinbaseInRelation) Row() []string {
	return []string{
		tx.BlockHash() + "_coinbase",
		tx.TxID(),
	}
}

type csvOutRelation struct {
	*types.Vout
	tx *types.Transaction
}

func (out csvOutRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (out csvOutRelation) Row() []string {
	return []string{
		out.tx.TxID(),
		out.tx.TxID() + "_" + strconv.FormatInt(int64(out.N()), 10),
	}
}

type csvLockedRelation struct {
	*types.Vout
	tx      *types.Transaction
	address string
}

func (out csvLockedRelation) Headers() []string {
	return []string{
		":START_ID",
		":END_ID",
	}
}

func (out csvLockedRelation) Row() []string {
	return []string{
		out.tx.TxID() + "_" + strconv.FormatInt(int64(out.N()), 10),
		out.address,
	}
}
