package postgres

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
)

func CreatePostgresTestServer(ctx context.Context) (testcontainers.Container, error) {
	env := make(map[string]string)
	env["POSTGRES_USER"] = "postgres"
	env["POSTGRES_PASSWORD"] = "postgres"

	const wd = "/Users/pearcek/Development/go/stock-indicators"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.3",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env:          env,
		Mounts: testcontainers.ContainerMounts{
			testcontainers.ContainerMount{
				Source: testcontainers.GenericBindMountSource{
					HostPath: fmt.Sprintf("%s/sql/init_db.sql", wd),
				},
				Target: "/docker-entrypoint-initdb.d/create_tables.sql",
			},
		},
	}
	postgresDBServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	return postgresDBServer, nil
}
