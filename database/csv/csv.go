package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
)

type CSVRecord interface {
	Headers() []string
	Row() []any
}

type CSVDatabase struct {
	csv_files  sync.Map
	maxWorkers int

	msgs chan CSVMsg

	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
	g        *errgroup.Group
}

func NewCSVDatabase(ctx context.Context, filePaths map[string]string, maxWorkers int) (*CSVDatabase, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no file paths provided")
	}
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	db := &CSVDatabase{
		maxWorkers: maxWorkers,
		ctx:        ctx,
		cancel:     cancel,
		g:          g,
	}
	for key, path := range filePaths {
		db.csv_files.Store(key, newCSVFile(path))
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		g.Go(db.csvWorker)
	}

	return db, nil
}

func (db *CSVDatabase) Close() error {
	done := false
	db.stopOnce.Do(func() {
		close(db.msgs)
		db.csv_files.Range(func(key, value interface{}) bool {
			csvFile := value.(*csvFile)
			// TODO: handle error
			csvFile.close()
			return true
		})
	})
	return db.g.Wait()
}

func (db *CSVDatabase) SendMsgAsync(msg CSVMsg) error {
	if db.ctx.Err() != nil {
		return db.ctx.Err()
	}
	db.msgs <- msg
	return nil
}

func (db *CSVDatabase) SendMsg(msg CSVMsg) error {
	if err := db.SendMsgAsync(msg); err != nil {
		return err
	}
	return msg.Wait()
}

func (db *CSVDatabase) SendMsgs(msgs []CSVMsg) error {
	for _, msg := range msgs {
		if err := db.SendMsgAsync(msg); err != nil {
			return err
		}
	}
	for _, msg := range msgs {
		if err := msg.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (db *CSVDatabase) insert(msg *CSVInsertMsg) error {
	val, ok := db.csv_files.Load(msg.FileKey())
	if !ok {
		return fmt.Errorf("csv file not found for %s", msg.FileKey())
	}
	if len(msg.records) == 0 {
		return nil
	}

	csvFile := val.(*csvFile)
	records := make([][]string, 0, len(msg.records)+1)
	if !csvFile.hasHeader {
		records = append(records, msg.records[0].Headers())
	}
	for _, record := range msg.records {
		records = append(records, anyRowToString(record.Row()))
	}
	if err := csvFile.writeAll(records); err != nil {
		return err
	}
	return nil
}

func (db *CSVDatabase) csvWorker() error {
	for msg := range db.msgs {
		switch msg := msg.(type) {
		case *CSVInsertMsg:
			msg.finish(db.insert(msg))
		default:
			return fmt.Errorf("unknown message type %T", msg)
		}
	}
	return nil
}
