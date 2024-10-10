package model_test

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

const dividendsTable = "dividends"

func TestDividendsSet_ToDB(t *testing.T) {

	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		t.Fatal(err.Error())
		return
	}

	ds, err := model.NewDividendsSetFromJSON(testDividendsData)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(len(ds.Dividends))

	err = ds.ToDB(context.Background(), pgxConn, dividendsTable)
	if err != nil {
		t.Error(err)
		return
	}

	responseDS := model.DividendsSet{}
	err = responseDS.FromDBbySymbol(context.Background(), pgxConn, dividendsTable, "CSX")
	if err != nil {
		t.Error(err)
		return
	}

	b, err := json.MarshalIndent(&responseDS, "", " ")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(b))

	if len(ds.Dividends) != len(responseDS.Dividends) {
		t.Error("Number of dividends does not match number of dividends")
	}

}
