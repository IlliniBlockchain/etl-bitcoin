package csv

import (
	"encoding/csv"
	"io"
	"math"
	"os"
	"sync"
)

type csvFile struct {
	filePath  string
	f         *os.File
	fstat     os.FileInfo
	mu        *sync.Mutex
	hasHeader bool

	// Cache file position of the last line read in csv for sequential reads.
	lastLineRead int
	lastLinePos  int64
}

func newCSVFile(filePath string) *csvFile {
	return &csvFile{
		filePath: filePath,
		mu:       &sync.Mutex{},
	}
}

func (cf *csvFile) file() (*os.File, error) {
	if cf.f != nil {
		return cf.f, nil
	}
	var err error
	cf.f, err = os.OpenFile(cf.filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	cf.fstat, err = cf.f.Stat()
	if err != nil {
		return nil, err
	}
	cf.hasHeader = cf.fstat.Size() > 0
	return cf.f, err
}

func (cf *csvFile) close() error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	if cf.f == nil {
		return nil
	}
	return cf.f.Close()
}

func (cf *csvFile) writeAll(records [][]string) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
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

func (cf *csvFile) read(line, limit int) ([][]string, error) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	f, err := cf.file()
	if err != nil {
		return nil, err
	}
	if line < 0 {
		line, err = cf.lineByOffset(line)
		if err != nil {
			return nil, err
		}
	}

	if err := cf.setLine(line); err != nil {
		return nil, err
	}
	f.Seek(cf.lastLinePos, 0)
	r := csv.NewReader(f)
	records := make([][]string, 0, limit)
	if limit <= 0 {
		limit = math.MaxInt
	} else {
		limit += line
	}

	// Skip header if it exists.
	if cf.hasHeader && cf.lastLineRead == 0 {
		r.Read()
		cf.lastLineRead++
	}
	for cf.lastLineRead < limit {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		records = append(records, record)
		cf.lastLineRead++
	}
	cf.lastLinePos += r.InputOffset()
	return records, nil
}

func (cf *csvFile) lineByOffset(offset int) (int, error) {
	cf.lastLineRead = 0
	cf.lastLinePos = 0
	if err := cf.setLine(math.MaxInt); err != nil {
		return 0, err
	}
	return cf.lastLineRead + offset, nil
}

func (cf *csvFile) setLine(line int) error {
	if cf.lastLineRead == line {
		return nil
	}
	f, _ := cf.file()
	if cf.lastLineRead > line {
		cf.lastLineRead = 0
		cf.lastLinePos = 0
	}
	f.Seek(cf.lastLinePos, 0)
	r := csv.NewReader(f)
	r.ReuseRecord = true
	for cf.lastLineRead < line {
		_, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		cf.lastLineRead++
	}
	cf.lastLinePos = r.InputOffset()
	return nil
}
