package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
)

type CSVDatabase struct {
	csv_files  sync.Map
	maxWorkers int

	msgs chan interface{}

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
	for name, path := range filePaths {
		csvFile, err := newCSVFile(path)
		if err != nil {
			return nil, err
		}
		db.csv_files.Store(name, csvFile)
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
		db.cancel()
		done = true
	})
	if !done {
		return fmt.Errorf("already closed")
	}
	return db.g.Wait()
}

func (db *CSVDatabase) SendMsg(msg interface{}) {
	db.msgs <- msg
}

func (db *CSVDatabase) csvWorker() error {
	for msg := range db.msgs {
		switch msg := msg.(type) {
		case *csvInsert:
			val, ok := db.csv_files.Load(msg.fileKey)
			if !ok {
				return fmt.Errorf("csv file not found for %s", msg.fileKey)
			}

			csvFile := val.(*csvFile)
			if csvFile.hasHeader {
				msg.records = msg.records[1:]
			}
			if err := csvFile.writeAll(msg.records); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown message type %T", msg)
		}
	}
	return nil
}

type csvFile struct {
	file      *os.File
	rwMutex   *sync.RWMutex
	hasHeader bool
}

func newCSVFile(filePath string) (*csvFile, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	fstat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &csvFile{
		file:      file,
		rwMutex:   &sync.RWMutex{},
		hasHeader: fstat.Size() > 0,
	}, nil
}

func (cf *csvFile) writeAll(records [][]string) error {
	cf.rwMutex.Lock()
	defer cf.rwMutex.Unlock()
	w := csv.NewWriter(cf.file)
	if err := w.WriteAll(records); err != nil {
		return err
	} else if err := w.Error(); err != nil {
		return err
	} else if !cf.hasHeader {
		cf.hasHeader = true
	}
	return nil
}

type csvInsert struct {
	fileKey string
	records [][]string
}
