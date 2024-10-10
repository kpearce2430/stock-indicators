package worksheets_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/xuri/excelize/v2"
	"testing"
)

const transactionWorkSheetFileName = "Transactions.xlsx"
const transactionWorkSheetName = "Transactions"

func TestWorkSheet_Transasctions(t *testing.T) {

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	w := worksheets.NewWorkSheet(excelize.NewFile(), pgxConn)
	w.Lookups = model.LoadLookupSet("1", string(lookups2))

	if err := w.Transactions(transactionWorkSheetName, utils.JulDate()); err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if err := w.File.DeleteSheet("Sheet1"); err != nil {
		t.Log(err.Error())
	}

	if err := w.File.SaveAs(transactionWorkSheetFileName); err != nil {
		t.Log(err)
		t.Fail()
	}
}
