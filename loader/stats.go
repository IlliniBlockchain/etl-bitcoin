package loader

import (
	"fmt"
	"time"

	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

type LoaderStats struct {
	blockRange      BlockRange
	start           time.Time
	loadingEnd      time.Time
	end             time.Time
	numBlocks       int64
	numTransactions int64
	done            chan struct{}
	dbTx            database.DBTx
	err             error
}

func (stats *LoaderStats) BlockRange() BlockRange { return stats.blockRange }

func (stats *LoaderStats) Start() time.Time { return stats.start }

func (stats *LoaderStats) LoadingEnd() time.Time { return stats.loadingEnd }

func (stats *LoaderStats) End() time.Time { return stats.end }

func (stats *LoaderStats) NumBlocks() int64 { return stats.numBlocks }

func (stats *LoaderStats) NumTransactions() int64 { return stats.numTransactions }

func (stats *LoaderStats) LoadingDuration() time.Duration {
	return stats.loadingEnd.Sub(stats.start)
}

func (stats *LoaderStats) TotalDuration() time.Duration {
	return stats.end.Sub(stats.start)
}

func (stats *LoaderStats) Wait() error {
	// TODO: safeguard against concurrent calls to Wait()
	defer close(stats.done)
	<-stats.done
	return stats.err
}

func (stats *LoaderStats) String() string {
	return fmt.Sprintf("Loaded %s (%d transactions) in %s and committed in %s", stats.blockRange, stats.numTransactions, stats.LoadingDuration(), stats.TotalDuration())
}

type dBTxWithStats struct {
	database.DBTx
	*LoaderStats
}

func newDBTxWithStats(dbTx database.DBTx, blockRange BlockRange) *dBTxWithStats {
	return &dBTxWithStats{
		DBTx: dbTx,
		LoaderStats: &LoaderStats{
			blockRange: blockRange,
			start:      time.Now(),
			done:       make(chan struct{}, 1),
			dbTx:       dbTx,
		},
	}
}

func (dbTx *dBTxWithStats) AddBlockHeader(blockHeader *types.BlockHeader) {
	dbTx.LoaderStats.numBlocks++
	dbTx.DBTx.AddBlockHeader(blockHeader)
}

func (dbTx *dBTxWithStats) AddTransaction(tx *types.Transaction) {
	dbTx.LoaderStats.numTransactions++
	dbTx.DBTx.AddTransaction(tx)
}

func (dbTx *dBTxWithStats) Commit() error {
	dbTx.LoaderStats.loadingEnd = time.Now()
	dbTx.err = dbTx.DBTx.Commit()
	dbTx.LoaderStats.end = time.Now()
	dbTx.LoaderStats.done <- struct{}{}
	return dbTx.err
}
