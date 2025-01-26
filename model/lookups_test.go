package model_test

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

func TestLoadLookupSet(t *testing.T) {
	t.Parallel()
	ls := model.LoadLookupSet("1", string(csvLookupData))
	if len(ls.LookUps) != 14 {
		t.Error("LookUp Count ", len(ls.LookUps), " does not equal 9")
	}
}

func TestLoadLookupToDB(t *testing.T) {
	const lookupTableName = "lookups"

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = model.LoadLookupFromCSV(context.TODO(), pgxConn, lookupTableName, csvLookupData)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ls, err := model.GetLookUpsFromDB(context.TODO(), pgxConn, lookupTableName)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(ls.LookUps) != 14 {
		t.Error("LookUp Count ", len(ls.LookUps), " does not equal 9")
	}
	t.Log(ls)
}
