package csv

import "sync"

// CSVMsg represents a message to be processed by a `CSVDatabase`.
type CSVMsg interface {
	// FileKey returns the key of the file to be processed.
	FileKey() string
	// Wait waits for the message to be processed.
	Wait() error
	// finish finishes the message with the given error.
	finish(err error)
}

// BaseCSVMsg is a base implementation of `CSVMsg`.
type BaseCSVMsg struct {
	fileKey  string
	err      error
	done     chan struct{}
	waitOnce sync.Once
}

// NewBaseCSVMsg creates a new `BaseCSVMsg`.
func NewBaseCSVMsg(fileKey string) *BaseCSVMsg {
	return &BaseCSVMsg{
		fileKey: fileKey,
		done:    make(chan struct{}, 1),
	}
}

// FileKey returns the key of the file to be processed.
//
// Implements `CSVMsg`.
func (msg *BaseCSVMsg) FileKey() string {
	return msg.fileKey
}

// Wait waits for the message to be processed.
//
// Implements `CSVMsg`.
func (msg *BaseCSVMsg) Wait() error {
	msg.waitOnce.Do(func() {
		<-msg.done
		close(msg.done)
	})
	return msg.err
}

func (msg *BaseCSVMsg) finish(err error) {
	msg.err = err
	msg.done <- struct{}{}
}

// CSVInsertMsg represents a message to insert records into a CSV file.
type CSVInsertMsg struct {
	// CSVMsg is the base message.
	CSVMsg
	// records are the records to be inserted.
	records []CSVRecord
}

// NewCSVInsertMsg creates a new `CSVInsertMsg`.
func NewCSVInsertMsg(fileKey string, records []CSVRecord) *CSVInsertMsg {
	return &CSVInsertMsg{
		CSVMsg:  NewBaseCSVMsg(fileKey),
		records: records,
	}
}

// CSVReadMsg represents a message to read records from a CSV file.
type CSVReadMsg struct {
	CSVMsg
	line    int
	limit   int
	Records [][]string
}

// NewCSVReadMsg creates a new `CSVReadMsg`.
func NewCSVReadMsg(fileKey string, line, limit int) *CSVReadMsg {
	return &CSVReadMsg{
		CSVMsg: NewBaseCSVMsg(fileKey),
		line:   line,
		limit:  limit,
	}
}
