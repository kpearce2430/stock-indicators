package worksheets

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

type ColumnInfo struct {
	File      *excelize.File
	Name      string
	SheetName string
	ColumnID  string
	maxSize   float64
	size      float64
	formula   bool
}

func NewColumnInfo(f *excelize.File, name, sheetName string, colNumber int) (*ColumnInfo, error) {
	colName, err := excelize.ColumnNumberToName(colNumber)
	if err != nil {
		return nil, err
	}
	ci := ColumnInfo{
		Name:      name,
		File:      f,
		SheetName: sheetName,
		ColumnID:  colName,
		maxSize:   0,
		size:      0,
	}
	return &ci, nil
}

func (ci *ColumnInfo) SetMaxSize(ms float64) *ColumnInfo {
	ci.maxSize = ms
	return ci
}

func (ci *ColumnInfo) SetSize(s float64) *ColumnInfo {
	ci.size = s
	return ci
}

func (ci *ColumnInfo) SetFormula(f bool) *ColumnInfo {
	ci.formula = f
	return ci
}

func (ci *ColumnInfo) WriteHeader(row int, style int) error {

	if float64(len(ci.Name)) > ci.size {
		ci.size = float64(len(ci.Name))
	}

	colRow := fmt.Sprintf("%s%d", ci.ColumnID, row)
	if err := ci.File.SetCellValue(ci.SheetName, colRow, ci.Name); err != nil {
		return err
	}
	if err := ci.File.SetCellStyle(ci.SheetName, colRow, colRow, style); err != nil {
		return err
	}
	return nil
}

func (ci *ColumnInfo) WriteCell(row int, value any, style int) error {

	v := fmt.Sprintf("%v", value)
	mySize := float64(len(v))
	switch {
	case ci.maxSize > 0:
		if mySize > ci.size {
			ci.size = mySize
		}
		if ci.size > ci.maxSize {
			ci.size = ci.maxSize
		}
	default:
		if mySize > ci.size {
			ci.size = mySize
		}
	}

	colRow := fmt.Sprintf("%s%d", ci.ColumnID, row)

	if ci.formula {
		if err := ci.File.SetCellFormula(ci.SheetName, colRow, colRow); err != nil {
			return err
		}
		if err := ci.File.SetCellFormula(ci.SheetName, colRow, fmt.Sprintf("%s", value)); err != nil {
			return err
		}
		if err := ci.File.SetCellStyle(ci.SheetName, colRow, colRow, style); err != nil {
			return err
		}
		return nil
	}

	if err := ci.File.SetCellValue(ci.SheetName, colRow, value); err != nil {
		return err
	}

	if err := ci.File.SetCellStyle(ci.SheetName, colRow, colRow, style); err != nil {
		return err
	}
	return nil
}

func (ci *ColumnInfo) GetColRow(row int) string {
	return fmt.Sprintf("%s%d", ci.ColumnID, row)
}

func (ci *ColumnInfo) SetColumnSize() error {
	if err := ci.File.SetColWidth(ci.SheetName, ci.ColumnID, ci.ColumnID, ci.size*1.4); err != nil {
		return err
	}
	return nil
}
