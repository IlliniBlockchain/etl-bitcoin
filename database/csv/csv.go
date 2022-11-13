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
	Row() []string
}

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

func (db *CSVDatabase) insert(msg *csvInsert) error {
	val, ok := db.csv_files.Load(msg.fileKey)
	if !ok {
		return fmt.Errorf("csv file not found for %s", msg.fileKey)
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
		records = append(records, record.Row())
	}
	if err := csvFile.writeAll(records); err != nil {
		return err
	}
	return nil
}

func (db *CSVDatabase) csvWorker() error {
	for msg := range db.msgs {
		switch msg := msg.(type) {
		case *csvInsert:
			if err := db.insert(msg); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown message type %T", msg)
		}
	}
	return nil
}

type csvFile struct {
	filePath  string
	f         *os.File
	fstat     os.FileInfo
	rwMutex   *sync.RWMutex
	hasHeader bool

	openOnce sync.Once
}

func newCSVFile(filePath string) *csvFile {
	return &csvFile{
		rwMutex: &sync.RWMutex{},
	}
}

func (cf *csvFile) file() (*os.File, error) {
	var err error
	cf.openOnce.Do(func() {
		cf.f, err = os.OpenFile(cf.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}

		cf.fstat, err = cf.f.Stat()
		if err != nil {
			return
		}
		cf.hasHeader = cf.fstat.Size() > 0
	})
	return cf.f, err
}

func (cf *csvFile) writeAll(records [][]string) error {
	cf.rwMutex.Lock()
	defer cf.rwMutex.Unlock()
	f, err := cf.file()
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		return err
	} else if err := w.Error(); err != nil {
		return err
	} else if !cf.hasHeader {
		cf.hasHeader = true
	}
	return nil
}

type csvMsg struct {
	fileKey string
}

type csvInsert struct {
	csvMsg
	records []CSVRecord
}

func newInsertMsg(fileKey string, records []CSVRecord) *csvInsert {
	return &csvInsert{
		csvMsg:  csvMsg{fileKey: fileKey},
		records: records,
	}
}