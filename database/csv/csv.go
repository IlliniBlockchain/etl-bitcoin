package csv

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// CSVRecord is a single row in a CSV file
type CSVRecord interface {
	Headers() []string
	Row() []string
}

// GetRowField returns the value of a field in a row.
func GetRowField(headers []string, row []string, key string) (string, error) {
	if len(headers) != len(row) {
		return "", fmt.Errorf("headers and row have different lengths")
	}
	for i, header := range headers {
		if header == key {
			return row[i], nil
		}
	}
	return "", fmt.Errorf("key not found")
}

// CSVDatabase is a base implementation of a database stored across CSV files.
type CSVDatabase struct {
	csv_files  sync.Map
	maxWorkers int
	msgs       chan CSVMsg
	ctx        context.Context
	stopOnce   sync.Once
	g          *errgroup.Group
}

// NewCSVDatabase creates a new CSVDatabase.
func NewCSVDatabase(ctx context.Context, filePaths map[string]string, maxWorkers int) (*CSVDatabase, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no file paths provided")
	}
	g, ctx := errgroup.WithContext(ctx)
	db := &CSVDatabase{
		maxWorkers: maxWorkers,
		msgs:       make(chan CSVMsg),
		ctx:        ctx,
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

// Close closes the database.
//
// Implements database.Database.
func (db *CSVDatabase) Close() error {
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

// SendMsgAsync sends a message to the database asynchronously.
func (db *CSVDatabase) SendMsgAsync(msg CSVMsg) {
	db.msgs <- msg
}

// SendMsg sends a message to the database synchronously.
func (db *CSVDatabase) SendMsg(msg CSVMsg) error {
	db.SendMsgAsync(msg)
	return msg.Wait()
}

// SendMsgs sends a batch of messages to the database synchronously.
func (db *CSVDatabase) SendMsgs(msgs []CSVMsg) error {
	for _, msg := range msgs {
		db.SendMsgAsync(msg)
	}
	for _, msg := range msgs {
		if err := msg.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (db *CSVDatabase) loadFile(fileKey string) (*csvFile, error) {
	val, ok := db.csv_files.Load(fileKey)
	if !ok {
		return nil, fmt.Errorf("csv file not found for %s", fileKey)
	}

	csvFile := val.(*csvFile)
	return csvFile, nil
}

func (db *CSVDatabase) insert(msg *CSVInsertMsg) error {
	csvFile, err := db.loadFile(msg.FileKey())
	if err != nil {
		return err
	}

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

func (db *CSVDatabase) read(msg *CSVReadMsg) error {
	csvFile, err := db.loadFile(msg.FileKey())
	if err != nil {
		return err
	}

	msg.Records, err = csvFile.read(msg.line, msg.limit)
	return err
}

func (db *CSVDatabase) csvWorker() error {
	for msg := range db.msgs {
		switch msg := msg.(type) {
		case *CSVInsertMsg:
			msg.finish(db.insert(msg))
		case *CSVReadMsg:
			msg.finish(db.read(msg))
		default:
			return fmt.Errorf("unknown message type %T", msg)
		}
	}
	return nil
}
