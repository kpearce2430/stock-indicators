package stocksheet

import (
	"errors"
	"slices"

	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/kpearce2430/stock-tools/stocksheet/styles"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// A file is like workbook made up of sheets
// A sheet is a like page made up of columns (or rows)
// A column is a set of data in common.

type StockFile struct {
	file   *excelize.File
	Styles *styles.Styles
	Sheets []StockSheet
}

type StockSheet struct {
	id      int
	name    string
	Columns []*column_info.ColumnInfo
}

var (
	ErrNoFileOpen         = errors.New("no file open")
	ErrNoSheetFound       = errors.New("no sheet found")
	ErrSheetAlreadyExists = errors.New("sheet already exists")
	ErrNoColumnFound      = errors.New("no column found")
)

func New() *StockFile {
	f := excelize.NewFile()
	if f == nil {
		logrus.Fatal("unable to create new excelize file")
		return nil
	}
	s, err := styles.New(f)
	if err != nil {
		logrus.Fatal("unable to create new styles ", err.Error())
		return nil
	}
	return NewStockFile(f, s)
}

func NewStockFile(f *excelize.File, s *styles.Styles) *StockFile {
	return &StockFile{
		file:   f,
		Styles: s,
		Sheets: []StockSheet{{
			id:   0,
			name: "Sheet1",
		}},
	}
}

func (sf *StockFile) GetExcelizeFile() *excelize.File {
	return sf.file
}

func (sf *StockFile) GetFile() *StockFile {
	return sf
}

func (sf *StockFile) GetStyles() *styles.Styles {
	return sf.Styles
}

func (sf *StockFile) GetSheet(name string) (*StockSheet, error) {
	for _, sheet := range sf.Sheets {
		if sheet.name == name {
			return &sheet, nil
		}
	}
	return nil, ErrNoSheetFound
}

func (sf *StockFile) SetFile(f *excelize.File) {
	sf.file = f
}

func (sf *StockFile) CloseFile() error {
	if sf.GetFile() == nil {
		return ErrNoFileOpen
	}
	return sf.file.Close()
}

func (sf *StockFile) Save(filename string) error {
	if sf.GetFile() == nil {
		return ErrNoFileOpen
	}
	return sf.file.SaveAs(filename)
}

func (sf *StockFile) DeleteSheet(name string) error {
	if sf.GetFile() == nil {
		return ErrNoFileOpen
	}
	for id, sheet := range sf.Sheets {
		if sheet.name == name {
			sf.Sheets = slices.Delete(sf.Sheets, id, id+1)
			return sf.file.DeleteSheet(name)
		}
	}
	return ErrNoSheetFound
}

func (sf *StockFile) NewSheet(name string) (*StockSheet, error) {
	sheet, err := sf.GetSheet(name)
	if err != nil && !errors.Is(err, ErrNoSheetFound) {
		return nil, err
	}
	if sheet != nil {
		return sheet, ErrSheetAlreadyExists
	}

	id, err := sf.file.NewSheet(name)
	if err != nil {
		logrus.Error("Error:", err.Error())
		return nil, err
	}

	sheet = &StockSheet{
		id:   id,
		name: name,
	}
	sf.Sheets = append(sf.Sheets, *sheet)
	return sheet, nil
}

func (sh *StockSheet) ID() int {
	return sh.id
}

func (sh *StockSheet) Name() string {
	return sh.name
}

func (sh *StockSheet) AddColumn(c *column_info.ColumnInfo) {
	sh.Columns = append(sh.Columns, c)
}

// GetColumn will return the columnInfo at the specific location in the list.
func (sh *StockSheet) GetColumn(index int) (*column_info.ColumnInfo, bool) {
	if index < 0 || index >= len(sh.Columns) {
		logrus.Error("Invalid column index:", index)
		return nil, false
	}
	return sh.Columns[index], true
}

func (sh *StockSheet) RemoveColumn(name string) error {
	for id, col := range sh.Columns {
		if col.Name == name {
			sh.Columns = slices.Delete(sh.Columns, id, id+1)
			return nil
		}
	}
	return ErrNoColumnFound
}
