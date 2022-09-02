package couchdb_client_test

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	couchdb_client "iex-indicators/couchdb-client"
	"log"
	"os"
	"testing"
)

var url string

func TestMain(m *testing.M) {
	log.Println("Do stuff BEFORE the tests!")

	env := make(map[string]string)

	env["COUCHDB_USER"] = "admin"
	env["COUCHDB_PASSWORD"] = "password"
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "couchdb-server:3.1.0",
		ExposedPorts: []string{"5984/tcp"},
		WaitingFor:   wait.ForListeningPort("5984/tcp"),
		Env:          env,
	}
	couchDBServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer couchDBServer.Terminate(ctx)

	ip, err := couchDBServer.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := couchDBServer.MappedPort(ctx, "5984")
	if err != nil {
		log.Fatal(err)
	}

	url = fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	log.Println(url)

	exitVal := m.Run()

	log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}

func TestCouchDBUp(t *testing.T) {

	client := couchdb_client.New(10, nil)
	if client.CouchDBClientValid() != true {
		t.Fatal("bad client")
	}

	if couchdb_client.CouchDBUp(url, &client) != true {
		t.Fatal("CouchDB Not Up")
	}

}
