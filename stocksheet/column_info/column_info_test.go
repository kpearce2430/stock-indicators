package column_info_test

import (
	"fmt"
	"github.com/kpearce2430/stock-tools/stocksheet/chart_builder"
	ci "github.com/kpearce2430/stock-tools/stocksheet/column_info"
	"github.com/kpearce2430/stock-tools/stocksheet/styles"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"math/rand"
	"testing"
	"time"
)

// TestColumnInfo is an initial test of the column_info class.
func TestColumnInfo(t *testing.T) {
	f := excelize.NewFile()
	const worksheetFile = "TestOne.xlsx"
	const worksheetName = "TestOne"
	s, err := styles.New(f)
	if err != nil {
		t.Error(err.Error())
		return
	}

	_, err = f.NewSheet(worksheetName)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ci, err := ci.New(f, worksheetName, worksheetName, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = ci.WriteHeader(1, s.HeaderStyle())
	if err != nil {
		t.Error(err.Error())
		return
	}

	if err = f.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = f.SaveAs(worksheetFile); err != nil {
		t.Error(err.Error())
		return
	}
}

// TestDefaultStyles tests the various styles.
func TestDefaultStyles(t *testing.T) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	styles, err := styles.New(f)
	if err != nil {
		t.Error(err.Error())
		return
	}

	colInfoName, err := ci.New(f, "Name", "Sheet1", 1)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = colInfoName.WriteHeader(1, styles.Header)
	if err != nil {
		t.Error(err.Error())
		return
	}

	colInfoValue, err := ci.New(f, "Value", "Sheet1", 2)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = colInfoValue.WriteHeader(1, styles.Header)
	assert.NoError(t, err, "Writing Value Header")

	type stylesTests struct {
		Name       string
		Value      any
		NameStyle  func(int) int
		ValueStyle func(int) int
	}

	tests := []stylesTests{
		{
			Name:       "Text",
			Value:      "Hello",
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.TextStyle,
		},
		{
			Name:       "Text",
			Value:      "World",
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.TextStyle,
		},
		{
			Name:       "Currency +",
			Value:      789.52,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.CurrencyStyle,
		},
		{
			Name:       "Currency -",
			Value:      -123.46,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.CurrencyStyle,
		},
		{
			Name:       "Date",
			Value:      time.Now(),
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.DateStyle,
		},
		{
			Name:       "Date",
			Value:      time.Now(),
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.DateStyle,
		},
		{
			Name:       "Number +",
			Value:      123.457,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.NumberStyle,
		},
		{
			Name:       "Number -",
			Value:      -642.35,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.NumberStyle,
		},
		{
			Name:       "Percent +",
			Value:      .055,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.PercentStyle,
		},
		{
			Name:       "Percent -",
			Value:      -.035,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.PercentStyle,
		},
		{
			Name:       "Accounting +",
			Value:      29,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.AccountingStyle,
		},
		{
			Name:       "Accounting -",
			Value:      -102.49,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.AccountingStyle,
		},
		{
			Name:       "Accounting dec",
			Value:      1702.49,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.AccountingStyle,
		},
		{
			Name:       "General - int",
			Value:      1,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.GeneralStyle,
		},
		{
			Name:       "General - float",
			Value:      1.99,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.GeneralStyle,
		},
		{
			Name:       "General - neg",
			Value:      -2.99,
			NameStyle:  styles.TextStyle,
			ValueStyle: styles.GeneralStyle,
		},
	}

	row := 2
	for i := 0; i < len(tests); i++ {
		err = colInfoName.WriteCell(row, tests[i].Name, tests[i].NameStyle(row))
		assert.NoError(t, err, "Error writing "+tests[i].Name)
		err = colInfoValue.WriteCell(row, tests[i].Value, tests[i].ValueStyle(row))
		assert.NoError(t, err, "Error writing value "+fmt.Sprintf("%d %v", i, tests[i].Value))
		row++
	}

	err = f.SaveAs("StylesTest.xlsx")
	if err != nil {
		t.Error(err.Error())
		return
	}

	_ = colInfoName.SetColumnSize()
	_ = colInfoValue.SetColumnSize()
}

func createWorksheet(worksheetName string) (*excelize.File, error) {
	//
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	_, err := f.NewSheet(worksheetName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// TestColumnInfo_AddComments tests Comments and a Formula.
func TestColumnInfo_AddComments(t *testing.T) {
	const worksheetFile = "ColumnInfoComments.xlsx"
	const worksheetComments = "Comments Sheet"

	f, err := createWorksheet(worksheetComments)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if f == nil {
		t.Error("excelize File is nil")
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	s, err := styles.New(f)
	if err != nil {
		t.Error(err.Error())
		return
	}

	row := 1
	colInfoName, err := ci.New(f, "Name", worksheetComments, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = colInfoName.WriteHeader(row, s.Header)
	if err != nil {
		t.Error(err.Error())
		return
	}

	colInfoValue, err := ci.New(f, "Value", worksheetComments, 2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = colInfoValue.WriteHeader(row, s.Header)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for i := range 12 {
		row++
		_ = colInfoName.WriteCell(row, time.Month(i+1), s.TextStyle(row))
		_ = colInfoName.AddComments(row, "kep", []string{"some comment ", fmt.Sprintf("%d", i)})
		_ = colInfoValue.WriteCell(row, rand.Float32(), s.NumberStyle(row))
	}
	row++
	_ = colInfoName.WriteCell(row, "Total", s.TextStyle(row))
	colInfoValue.SetFormula(true)
	_ = colInfoValue.WriteCell(row, "=sum(b2:b13)", s.NumberStyle(row))

	if err = f.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = f.SaveAs(worksheetFile); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestColumnInfo_ChartBuilder(t *testing.T) {
	const worksheetFile = "ColumnInfoChart.xlsx"
	const worksheetName = "Column Chart"

	f, err := createWorksheet(worksheetName)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if f == nil {
		t.Error("excelize file is nil")
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	styles, err := styles.New(f)
	if err != nil {
		t.Error(err.Error())
		return
	}

	row := 1
	colInfoName, err := ci.New(f, "Name", worksheetName, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = colInfoName.WriteHeader(row, styles.Header)
	if err != nil {
		t.Error(err.Error())
		return
	}

	colInfoValue, err := ci.New(f, "Value", worksheetName, 2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = colInfoValue.WriteHeader(row, styles.Header)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for i := range 12 {
		row++
		_ = colInfoName.WriteCell(row, time.Month(i+1), styles.TextStyle(row))
		_ = colInfoName.AddComments(row, "kep", []string{"some comment ", fmt.Sprintf("%d", i)})
		_ = colInfoValue.WriteCell(row, rand.Float32(), styles.NumberStyle(row))
	}
	row++
	_ = colInfoName.WriteCell(row, "Total", styles.TextStyle(row))
	colInfoValue.SetFormula(true)
	_ = colInfoValue.WriteCell(row, "=sum(b2:b13)", styles.NumberStyle(row))

	chartTest := chart_builder.ChartBuilder{
		File:          f,
		WorksheetName: worksheetName,
		Title:         "Year Over Year Dividends",
		Type:          excelize.Col,
		Height:        250,
		Width:         950,
		VaryColors:    true,
	}

	chartTest.AddCategorySeries(colInfoName.ColumnID, 2, colInfoName.ColumnID, 13)
	chartTest.AddValueSeries(colInfoValue.ColumnID, 2, colInfoValue.ColumnID, 13)
	if err = chartTest.BuildChart("d3"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = f.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = f.SaveAs(worksheetFile); err != nil {
		t.Error(err.Error())
		return

	}
}
