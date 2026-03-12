package monte_carlo

import (
	"fmt"

	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/kpearce2430/stock-tools/stocksheet/styles"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type MonteCarloInterface interface {
	// GetStockSheet() (*stocksheet.StockSheet,error)
	GetFile() *stocksheet.StockFile
	GetStyles() *styles.Styles
	GetExcelizeFile() *excelize.File
}

type MonteCarlo struct {
	M worksheets.WorksheetInterface
}

func NewMonteCarlo(m MonteCarloInterface) *MonteCarlo {
	return &MonteCarlo{M: m}
}

func (m *MonteCarlo) monteCarloHeaders(sheet *stocksheet.StockSheet, worksheetName string) error {
	column := 1
	row := 1
	const years = 9

	colInfo, err := column_info.New(m.M.GetExcelizeFile(), "Control", worksheetName, column)
	if err != nil {
		return err
	}

	sheet.AddColumn(colInfo)

	for i := range years {
		column++
		colInfo, err = column_info.New(m.M.GetExcelizeFile(), fmt.Sprintf("Year %d", i+1), worksheetName, column)
		if err != nil {
			return err
		}
		colInfo.SetFormula(true)
		sheet.AddColumn(colInfo)
		column++

		colInfo, err = column_info.New(m.M.GetExcelizeFile(), fmt.Sprintf("Change %d", i+1), worksheetName, column)
		if err != nil {
			return err
		}
		colInfo.SetFormula(true)
		sheet.AddColumn(colInfo)
	}

	colInfo, err = column_info.New(m.M.GetExcelizeFile(), fmt.Sprintf("Year %d", years), worksheetName, column)
	if err != nil {
		return err
	}
	if colInfo == nil {
		logrus.Fatal("uninitialized column info")
	}

	colInfo.SetFormula(true)
	sheet.AddColumn(colInfo)

	for _, colInfo = range sheet.Columns {
		_ = colInfo.WriteHeader(row, m.M.GetStyles().Header)
	}
	return nil
}

func (m *MonteCarlo) monteCarloRowCol0(row int, colInfo *column_info.ColumnInfo) error {
	switch row {
	case 0, 1:
		return nil
	case 2: // Starting Balance
		_ = colInfo.WriteCell(row, "Start", m.M.GetStyles().TextStyle(row))
	case 3:
		_ = colInfo.WriteCell(row, 1000000, m.M.GetStyles().TextStyle(row))
	case 4: // Mean
		_ = colInfo.WriteCell(row, "Mean", m.M.GetStyles().TextStyle(row))
	case 5:
		_ = colInfo.WriteCell(row, .03, m.M.GetStyles().TextStyle(row))
	case 6:
		_ = colInfo.WriteCell(row, "StDev", m.M.GetStyles().TextStyle(row))
	case 7:
		_ = colInfo.WriteCell(row, .10, m.M.GetStyles().TextStyle(row))
	case 8: // Average
		_ = colInfo.WriteCell(row, "Average", m.M.GetStyles().TextStyle(row))
	case 9:
		_ = colInfo.WriteCell(row, "=average(t2:v10000)", m.M.GetStyles().AccountingStyle(row))
	case 10: // Min
		_ = colInfo.WriteCell(row, "Min", m.M.GetStyles().TextStyle(row))
	case 11:
		_ = colInfo.WriteCell(row, "=min(t2:v10000)", m.M.GetStyles().AccountingStyle(row))
	case 12: // Max
		_ = colInfo.WriteCell(row, "Max", m.M.GetStyles().TextStyle(row))
	case 13:
		_ = colInfo.WriteCell(row, "=max(t2:v10000)", m.M.GetStyles().AccountingStyle(row))
	case 14: // StdDv.P
		_ = colInfo.WriteCell(row, "Max", m.M.GetStyles().TextStyle(row))
	case 15:
		_ = colInfo.WriteCell(row, "=stdev.p(v2:v10001)", m.M.GetStyles().AccountingStyle(row))
	default:
		// noop
	}
	return nil
}

func (m *MonteCarlo) monteCarloRow(columns []*column_info.ColumnInfo, row int) error {
	//
	if row < 2 {
		return nil
	}
	for i, colInfo := range columns {
		switch i {
		case 0:
			_ = m.monteCarloRowCol0(row, columns[0])
		case 1:
			_ = colInfo.WriteCell(row, "=$a$3", m.M.GetStyles().AccountingStyle(row))
		case 2, 4, 6, 8, 10, 12, 14, 16, 18, 20:
			//TODO: formula isn't being written correctly.
			_ = colInfo.WriteCell(row, "=NORM.INV(RAND(), $A$5, $A$7)", m.M.GetStyles().NumberStyle(row))
		default:
			// =B2+(B2*C2)
			prevCol := columns[i-2]
			prevColumnID := prevCol.ColumnID

			prevForm := columns[i-1]
			prevFormID := prevForm.ColumnID
			formula := fmt.Sprintf("=$%s%d+($%s%d*%s%d)-(%s%d*.04)",
				prevColumnID, row,
				prevColumnID, row,
				prevFormID, row,
				prevColumnID, row)

			_ = colInfo.WriteCell(row, formula, m.M.GetStyles().AccountingStyle(row))
		}
	}
	return nil
}

// MonteCarlo generates a Monte Carlo simulation for stock performance.  This is still a work in progress.
func (m *MonteCarlo) MonteCarlo(worksheetName string) error {
	stockFile := m.M.GetFile()
	sheet, err := stockFile.NewSheet(worksheetName)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	if err = m.monteCarloHeaders(sheet, worksheetName); err != nil {
		logrus.Error(err.Error())
		return err
	}

	// =NORM.INV(RAND(), $A$2, $A$3)
	for row := range 10001 {
		err = m.monteCarloRow(sheet.Columns, row)
		if err != nil {
			return err
		}
	}

	for _, colInfo := range sheet.Columns {
		_ = colInfo.SetColumnSize()
	}
	return nil
}
