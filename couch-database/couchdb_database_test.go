package couch_database_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	couch_database "iex-indicators/couch-database"
	"log"
	"os"
	"testing"
)

type TestDocument struct {
	Id    string `json:"_id,omitempty"`
	Rev   string `json:"_rev,omitempty"`
	Name  string
	Value int64
}

var url string

func TestMain(m *testing.M) {

	ctx := context.Background()

	couchDBServer, _ := couch_database.CreateCouchDBServer(ctx)
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

	os.Setenv("DATABASE_NAME", "tester")
	os.Setenv("COUCHDB_URL", url)
	os.Setenv("COUCHDB_USER", "admin")
	os.Setenv("COUCHDB_PASSWORD", "password")

	m.Run()

}

func TestDatabaseConfig(t *testing.T) {

	dbConfig, err := couch_database.NewDatabaseConfig("")

	assert.Nil(t, err, fmt.Sprintf("%+v", err))

	assert.Equal(t, "tester", dbConfig.DatabaseName, "database name mismatch")
	assert.Equal(t, url, dbConfig.CouchDBUrl, "database url mismatch")
	assert.Equal(t, "admin", dbConfig.Username, "couchdb username mismatch")
	assert.Equal(t, "password", dbConfig.Password, "couchdb password mismatch")

	databaseStore := couch_database.NewDataStore[TestDocument](dbConfig)

	assert.NotNil(t, databaseStore, "database store is nil")

	if databaseStore.DatabaseCreate() != true {
		t.Fatal("Error creating a database")
	}

	log.Printf("Database created")

	testDocument := TestDocument{Name: "name", Value: 1}

	_, err = databaseStore.DocumentCreate("key", &testDocument)

	if err != nil {
		t.Fatal(err)
	}

}

func TestDataStore(t *testing.T) {

	err := os.Setenv("MY_DATABASE_NAME", "junk")
	if err != nil {
		return
	}
	os.Setenv("MY_COUCHDB_URL", url)
	os.Setenv("MY_COUCHDB_USER", "admin")
	os.Setenv("MY_COUCHDB_PASSWORD", "password")

	databaseStore := couch_database.DataStore[TestDocument]("MY")

	assert.NotNil(t, databaseStore, "database store is nil")

	if databaseStore.DatabaseCreate() != true {
		t.Fatal("Error creating a database")
	}

	log.Printf("Database created")

	testDocument := TestDocument{Name: "name", Value: 1}

	_, err = databaseStore.DocumentCreate("key", &testDocument)

	if err != nil {
		t.Fatal(err)
	}

}

func TestDatabaseStore_CouchDBUp(t *testing.T) {

	databaseStore := couch_database.New[TestDocument]("name", url, "admin", "password")
	if databaseStore.CouchDBUp() == true {
		log.Println("Datastore Couch DB is Up")
	} else {
		t.Fatal("Couchdb not up")
	}
}
func TestCouchDBUp(t *testing.T) {

	databaseStore := couch_database.New[TestDocument]("name", url, "admin", "password")

	if databaseStore.DatabaseCreate() != true {
		t.Fatal("Error creating a database")
	}

	log.Printf("Database created")

	testDocument := TestDocument{Name: "name", Value: 1}

	revision, err := databaseStore.DocumentCreate("key", &testDocument)

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("Document created revision: %s", revision)

	getDocument, err := databaseStore.DocumentGet("key")

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%s, %s, %s, %d", getDocument.Id, getDocument.Rev, getDocument.Name, getDocument.Value)

	getDocument.Name = "New Name"

	revision, err = databaseStore.DocumentUpdate(getDocument.Id, getDocument.Rev, getDocument)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Document updated new revision: %s", revision)

	getDocument, err = databaseStore.DocumentGet("key")

	if err != nil {
		t.Fatal(err)
	}

	couchDatabaseInfo, err := databaseStore.DatabaseExists()
	assert.Nil(t, err, "Database Exists returns error")
	t.Log("db info>", couchDatabaseInfo)

	log.Printf("{ %s, %s, %s, %d }", getDocument.Id, getDocument.Rev, getDocument.Name, getDocument.Value)

	revision, err = databaseStore.DocumentDelete("key", getDocument.Rev)

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("Document deleted new revision: %s", revision)

	t.Log("all done")

}
