package loader

import "github.com/IlliniBlockchain/etl-bitcoin/types"

// Loader represents loading process from a bitcoin client.
type Loader interface {
	// export all
	ExportAll(blockExportJob, txExportJob ExportJob)
	// export blocks (headers?)
	ExportBlocks(blockExportJob ExportJob)
	// export transactions
	ExportTransactions(txExportJob ExportJob)
}

type ExportJob struct {
	exportDir      string
	startBlockNum  int64
	endBlockNum    int64
	maxRowsPerFile int32
}

func BlockToRecord(*types.Block) []string {
	return nil
}

func BlockHeaderToRecord(*types.BlockHeader) []string {
	return nil
}

func TxToRecord(*types.Transaction) []string {
	return nil
}
