package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed sql/init_db.sql
var initDB string

func ConnectToPostgres() (*pgxpool.Pool, error) {
	pgxConn, err := pgxpool.New(context.Background(), utils.GetEnv("PG_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres"))
	if err != nil {
		return nil, err
	}

	if pgxConn == nil {
		return nil, errors.New("nil connection")
	}
	return pgxConn, nil
}

func CreatePostgresTestServer(ctx context.Context) (testcontainers.Container, error) {
	env := make(map[string]string)
	env["POSTGRES_USER"] = "postgres"
	env["POSTGRES_PASSWORD"] = "postgres"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.3",
		Name:         "postgres-test",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env:          env,
		FromDockerfile: testcontainers.FromDockerfile{
			Dockerfile: "Dockerfile",
		},
	}
	postgresDBServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})

	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	pgIP, err := postgresDBServer.Host(ctx)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	pgMappedPort, err := postgresDBServer.MappedPort(ctx, "5432")
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	// postgres://postgres:postgres@localhost:5432/postgres
	pgURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres", pgIP, pgMappedPort.Port())
	err = os.Setenv("PG_DATABASE_URL", pgURL)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	if err = createStockTables(ctx); err != nil {
		logrus.Fatal(err.Error())
	}

	return postgresDBServer, nil
}

func createStockTables(ctx context.Context) error {
	pgxConn, err := ConnectToPostgres()
	if err != nil {
		return err
	}

	_, err = pgxConn.Exec(ctx, initDB)
	if err != nil {
		return err
	}
	return nil
}
