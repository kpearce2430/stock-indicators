package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/postgres"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresDBServer, _ := postgres.CreatePostgresTestServer(ctx)
	defer func() {
		_ = postgresDBServer.Terminate(ctx)
	}()

	os.Exit(m.Run())
}

func TestPostgres_New(t *testing.T) {
	t.Log("Test Postgres New")
	ctx := context.Background()

	pgxConn, err := pgxpool.New(ctx, os.Getenv("PG_DATABASE_URL"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	selectStatement := `
SELECT table_name
FROM information_schema.tables
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
AND table_type = 'BASE TABLE';
`

	rows, err := pgxConn.Query(ctx, selectStatement)
	defer rows.Close()
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	var tables []string
	// Iterate through the result set
	num := 0
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			logrus.Fatal(err.Error())
		}
		tables = append(tables, table)
		num++
	}

	t.Log(tables)

}
