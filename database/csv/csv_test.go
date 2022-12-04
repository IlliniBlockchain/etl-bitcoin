package csv

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type invalidMsg struct{}

func (m *invalidMsg) FileKey() string  { return "" }
func (m *invalidMsg) Wait() error      { return nil }
func (m *invalidMsg) finish(err error) {}

type testRow struct {
	header1 int
	header2 string
}

func (t *testRow) Headers() []string {
	return []string{"header1", "header2"}
}

func (t *testRow) Row() []string {
	return []string{strconv.Itoa(t.header1), t.header2}
}

var (
	testRow1 = &testRow{
		header1: 1,
		header2: "test1",
	}
	testRow2 = &testRow{
		header1: 2,
		header2: "test2",
	}
)

func TestNewCSVDatabase(t *testing.T) {
	type args struct {
		filePaths  map[string]string
		maxWorkers int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				filePaths: map[string]string{
					"test": "test.csv",
				},
				maxWorkers: 1,
			},
			wantErr: false,
		},
		{
			name: "no_file_paths",
			args: args{
				filePaths:  nil,
				maxWorkers: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCSVDatabase(context.Background(), tt.args.filePaths, tt.args.maxWorkers)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCSVDatabaseInvalidMsg(t *testing.T) {
	db, err := NewCSVDatabase(context.Background(), map[string]string{"test": "test.csv"}, 1)
	assert.NoError(t, err)
	msg := &invalidMsg{}
	db.SendMsg(msg)
	assert.Error(t, db.Close())
}

type CSVDatabaseTestSuite struct {
	suite.Suite
	db        *CSVDatabase
	filePaths map[string]string
}

func (s *CSVDatabaseTestSuite) SetupTest() {
	var err error
	s.filePaths = map[string]string{
		"basic":             "./testdata/basic.csv",
		"basic_test":        "./testdata/basic_test.csv",
		"invalid_file_path": "",
	}
	s.db, err = NewCSVDatabase(context.Background(), s.filePaths, 1)
	assert.NoError(s.T(), err)
}

func (s *CSVDatabaseTestSuite) TearDownTest() {
	assert.NoError(s.T(), s.db.Close())
	for _, filePath := range s.filePaths {
		if strings.Contains(filePath, "_test.csv") {
			os.Remove(filePath)
		}
	}
}

func (s *CSVDatabaseTestSuite) TestInsert() {
	type args struct {
		table string
		rows  []CSVRecord
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				table: "basic_test",
				rows:  []CSVRecord{testRow1, testRow2},
			},
			wantErr: false,
		},
		{
			name: "no_table",
			args: args{
				table: "no_table",
				rows:  []CSVRecord{testRow1},
			},
			wantErr: true,
		},
		{
			name: "invalid_file_path",
			args: args{
				table: "invalid_file_path",
				rows:  []CSVRecord{testRow1},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			msg := NewCSVInsertMsg(tt.args.table, tt.args.rows)
			err := s.db.SendMsg(msg)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				haveFile := s.filePaths[tt.args.table]
				wantFile := s.filePaths[strings.ReplaceAll(tt.args.table, "_test", "")]
				haveFileBytes, err := os.ReadFile(haveFile)
				s.NoError(err)
				wantFileBytes, err := os.ReadFile(wantFile)
				s.NoError(err)
				s.Equal(string(wantFileBytes), string(haveFileBytes))
			}
		})
	}
}

func (s *CSVDatabaseTestSuite) TestRead() {
	type args struct {
		table string
		line  int
		limit int
	}
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				table: "basic",
			},
			want: [][]string{
				testRow1.Row(),
				testRow2.Row(),
			},
			wantErr: false,
		},
		{
			name: "success_by_offset_1",
			args: args{
				table: "basic",
				line:  -1,
				limit: 1,
			},
			want: [][]string{
				testRow2.Row(),
			},
			wantErr: false,
		},
		{
			name: "success_by_offset_2",
			args: args{
				table: "basic",
				line:  -2,
				limit: 1,
			},
			want: [][]string{
				testRow1.Row(),
			},
			wantErr: false,
		},
		{
			name: "no_table",
			args: args{
				table: "no_table",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid_file_path",
			args: args{
				table: "invalid_file_path",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			msg := NewCSVReadMsg(tt.args.table, tt.args.line, tt.args.limit)
			err := s.db.SendMsg(msg)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.want, msg.Records)
			}
		})
	}
}

func (s *CSVDatabaseTestSuite) TestSequentialRead() {
	msg1 := NewCSVReadMsg("basic", 1, 1)
	msg2 := NewCSVReadMsg("basic", 2, 1)
	err := s.db.SendMsgs([]CSVMsg{msg1, msg2})
	s.NoError(err)
	s.Equal([][]string{testRow1.Row()}, msg1.Records)
	s.Equal([][]string{testRow2.Row()}, msg2.Records)
}

func TestCSVDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(CSVDatabaseTestSuite))
}

func TestGetRowField(t *testing.T) {
	type args struct {
		headers []string
		row     []string
		key     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "first_field",
			args: args{
				headers: testRow1.Headers(),
				row:     testRow1.Row(),
				key:     "header1",
			},
			want: "1",
		},
		{
			name: "second_field",
			args: args{
				headers: testRow1.Headers(),
				row:     testRow1.Row(),
				key:     "header2",
			},
			want: "test1",
		},
		{
			name: "field_not_found",
			args: args{
				headers: testRow1.Headers(),
				row:     testRow1.Row(),
				key:     "header3",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRowField(tt.args.headers, tt.args.row, tt.args.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func BenchmarkCSVDatabaseSequentialRead(b *testing.B) {
	bmFilePath := fmt.Sprintf("./testdata/%s%d.csv", b.Name(), b.N)
	db, err := NewCSVDatabase(context.Background(), map[string]string{"bm": bmFilePath}, 1)
	if err != nil {
		b.Fatal(err)
	}
	records := make([]CSVRecord, b.N)
	msgs := make([]CSVMsg, b.N)
	for i := 0; i < b.N; i++ {
		records[i] = &testRow{
			i + 1,
			"test%d" + strconv.Itoa(i+1),
		}
		msgs[i] = NewCSVReadMsg("bm", i+1, 1)
	}
	if err := db.SendMsg(NewCSVInsertMsg("bm", records)); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	db.SendMsgs(msgs)
	b.StopTimer()
	if err := db.Close(); err != nil {
		b.Fatal(err)
	}
	if err := os.Remove(bmFilePath); err != nil {
		b.Fatal(err)
	}
}
