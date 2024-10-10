package worksheets_test

import (
	"fmt"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
	"testing"
	"time"
)

func TestDefaultStyles(t *testing.T) {

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	styles, err := worksheets.DefaultStyles(f)
	assert.NoError(t, err, "Creating DefaultStyles")
	colInfoName, err := worksheets.NewColumnInfo(f, "Name", "Sheet1", 1)
	assert.NoError(t, err, "Creating NewColumnInfo")
	err = colInfoName.WriteHeader(1, styles.Header)
	assert.NoError(t, err, "Writing Name Header")

	colInfoValue, err := worksheets.NewColumnInfo(f, "Value", "Sheet1", 2)
	assert.NoError(t, err, "Creating NewColumnInfo")
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
	assert.NoError(t, err, "Saving StylesText.xlxs")

	_ = colInfoName.SetColumnSize()
	_ = colInfoValue.SetColumnSize()
}
