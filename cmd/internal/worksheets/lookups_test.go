package worksheets_test

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

func TestLookups(t *testing.T) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)

	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = model.LoadLookupSet("1", string(lookups2))
	if err = w.LookupSheet("Lookups"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = w.StockFile.DeleteSheet("Sheet1"); err != nil {
		t.Error(err.Error())
		return
	}

	if err = w.StockFile.Save("Lookups.xlsx"); err != nil {
		t.Error(err.Error())
		return
	}
}
