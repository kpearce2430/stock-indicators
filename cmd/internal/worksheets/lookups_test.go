package worksheets_test

import (
	"context"
	_ "embed"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"strings"
	"testing"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

//go:embed testdata/portfolio_value.csv
var csvPortfolioValueData []byte

func testWorksheet(f *excelize.File, styles *worksheets.Styles, worksheetName string, rawData []byte) {
	var columns []*worksheets.ColumnInfo
	r := csv.NewReader(strings.NewReader(string(rawData)))
	f.NewSheet(worksheetName)

	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if i == 0 {
			for j := 0; j < len(record); j++ {
				ci, err := worksheets.NewColumnInfo(f, record[j], worksheetName, j+1)
				if err != nil {
					log.Fatal(err.Error())
				}
				columns = append(columns, ci)
				ci.WriteHeader(i+1, styles.Header)
			}
			continue
		}

		// Name,Symbol,Shares,Type,Price,Price Day Change,Price Day Change (%),Cost Basis,Market Value,Average Cost Per Share,Gain/Loss 12-Month,Gain/Loss,Gain/Loss (%)
		for j := 0; j < len(record); j++ {
			ci := columns[j]
			switch ci.Name {
			case "Name", "Symbol", "Type":
				ci.WriteCell(i+1, record[j], styles.TextStyle(i))
			case "Shares":
				value, err := utils.FloatParse(record[j])
				if err != nil {
					ci.WriteCell(i+1, record[j], styles.TextStyle(i))
					continue
				}
				ci.WriteCell(i+1, value, styles.NumberStyle(i))
			case "Price", "Price Day Change", "Cost Basis", "Market Value", "Average Cost Per Share":
				value, err := utils.FloatParse(record[j])
				if err != nil {
					ci.WriteCell(i+1, record[j], styles.TextStyle(i))
					continue
				}
				ci.WriteCell(i+1, value, styles.CurrencyStyle(i))
			case "Price Day Change (%)", "Gain/Loss (%)":
				value, err := utils.FloatParse(record[j])
				if err != nil {
					ci.WriteCell(i+1, record[j], styles.TextStyle(i))
					continue
				}
				ci.WriteCell(i+1, value/100.0, styles.PercentStyle(i))
			case "Gain/Loss 12-Month", "Gain/Loss":
				value, err := utils.FloatParse(record[j])
				if err != nil {
					ci.WriteCell(i+1, record[j], styles.TextStyle(i))
					continue
				}
				ci.WriteCell(i+1, value, styles.AccountingStyle(i))

			default:
				ci.WriteCell(i+1, record[j], styles.TextStyle(i))
			}
		}
	}

	for _, c := range columns {
		_ = c.SetColumnSize()
	}
}

func TestLookups(t *testing.T) {

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	w := worksheets.NewWorkSheet(excelize.NewFile(), pgxConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	if err := w.LookupSheet("Lookups"); err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if err := w.File.SaveAs("lookups.xlsx"); err != nil {
		t.Log(err)
		t.Fail()
	}
}
