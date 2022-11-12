package csv

import "sync"

type CSVMsg interface {
	FileKey() string
	Wait() error
	finish(err error)
}

type BaseCSVMsg struct {
	fileKey  string
	err      error
	done     chan struct{}
	waitOnce sync.Once
}

func NewBaseCSVMsg(fileKey string) *BaseCSVMsg {
	return &BaseCSVMsg{
		fileKey: fileKey,
		done:    make(chan struct{}),
	}
}

func (msg *BaseCSVMsg) FileKey() string {
	return msg.fileKey
}

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

type CSVInsertMsg struct {
	CSVMsg
	records []CSVRecord
}

func NewCSVInsertMsg(fileKey string, records []CSVRecord) *CSVInsertMsg {
	return &CSVInsertMsg{
		CSVMsg:  NewBaseCSVMsg(fileKey),
		records: records,
	}
}
