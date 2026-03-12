package lookups_test

import (
	"context"
	_ "embed"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	lookupsworksheet "github.com/kpearce2430/stock-tools/cmd/internal/worksheets/lookups"
	lookupsmodel "github.com/kpearce2430/stock-tools/model/lookups"

	"testing"
)

var (
	//go:embed testdata/lookups.csv
	lookupCSV string
)

var ls *lookupsmodel.LookUpSet

func TestMain(m *testing.M) {
	ls = lookupsmodel.LoadLookupSet("1", lookupCSV)
	os.Exit(m.Run())
}

func TestLookups(t *testing.T) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	w := worksheets.New(pgxConn)
	l := lookupsworksheet.New(w)

	defer func() {
		_ = w.StockFile.CloseFile()
	}()

	w.Lookups = ls
	if err = l.LookupSheet("Lookups"); err != nil {
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
