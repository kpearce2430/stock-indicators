package couch_database

import (
	"context"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/kelseyhightower/envconfig"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	couchdb_client "iex-indicators/couchdb-client"
	"iex-indicators/http-client"
	"log"
)

type DatabaseStore[T interface{}] struct {
	databaseConfig *DatabaseConfig
	httpClient     *req.Client
}

func GetDataStoreByDatabaseName[T interface{}](databaseName string) (*DatabaseStore[T], error) {

	var dbConfig DatabaseConfig
	err := envconfig.Process("", &dbConfig)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	// log.Println("Config>>", dbConfig.Username, dbConfig.CouchDBUrl)
	dbConfig.DatabaseName = databaseName

	datastore := NewDataStore[T](&dbConfig)

	return &datastore, nil

}
func NewDataStore[T interface{}](config *DatabaseConfig) DatabaseStore[T] {

	return DatabaseStore[T]{config, http_client.GetDefaultClient(10, false)}
}

func DataStore[T interface{}](prefix string) DatabaseStore[T] {

	dbConfig, _ := NewDatabaseConfig(prefix)
	dbStore := NewDataStore[T](dbConfig)
	return dbStore

}

func New[T interface{}](name string, url string, user string, pswd string) DatabaseStore[T] {

	dbConfig := DatabaseConfig{name, url, user, pswd}
	dbStore := DatabaseStore[T]{&dbConfig, http_client.GetDefaultClient(10, false)}
	return dbStore

}

// CreateCouchDBServer I put this here so that other test packages can use it.
func CreateCouchDBServer(ctx context.Context) (testcontainers.Container, error) {
	env := make(map[string]string)

	env["COUCHDB_USER"] = "admin"
	env["COUCHDB_PASSWORD"] = "password"

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

	return couchDBServer, nil

}

func (ds DatabaseStore[T]) CouchDBUp() bool {

	url := fmt.Sprintf("%s/_up", ds.databaseConfig.CouchDBUrl)
	var result couchdb_client.CouchDBStatus

	ds.httpClient.R().
		SetResult(&result).
		Get(url)

	// return couchdb_client.CouchDBUp(ds.couchDBUrl, &ds.httpClient)
	if result.Status == "ok" {
		return true
	}
	return false
}

func (ds DatabaseStore[T]) GetConfig() string {

	return fmt.Sprintf("%s : %s : %s :%s",
		ds.databaseConfig.DatabaseName, ds.databaseConfig.CouchDBUrl,
		ds.databaseConfig.Username, ds.databaseConfig.Password)
}

func (ds DatabaseStore[T]) DatabaseExists() (*CouchDatabaseInfo, error) {

	var couchDatabaseInfo CouchDatabaseInfo
	createDatabaseURL := fmt.Sprintf("%s/%s", ds.databaseConfig.CouchDBUrl, ds.databaseConfig.DatabaseName)
	// fmt.Println("url>>", createDatabaseURL)

	resp, err := ds.httpClient.R().
		SetResult(&couchDatabaseInfo).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		// SetError(&err).
		Get(createDatabaseURL)

	if err != nil {
		fmt.Println("Get>>", err.Error())
		return nil, err
	}

	if resp.IsError() {
		// fmt.Printf("Is Error>> Database %s returns %v", ds.GetConfig(), resp.Status)
		err := fmt.Errorf("Database %s returns %v", ds.GetConfig(), resp.Status)
		return nil, err
	}

	// log.Println("resp>", resp.Status)
	// log.Println("data>>", couchDatabaseInfo)

	return &couchDatabaseInfo, nil

}

func (ds DatabaseStore[T]) DatabaseCreate() bool {

	createDatabaseURL := fmt.Sprintf("%s/%s", ds.databaseConfig.CouchDBUrl, ds.databaseConfig.DatabaseName)

	var couchDBResponse couchdb_client.CouchDBResponse
	//	var err error

	resp, err := ds.httpClient.R().
		SetResult(&couchDBResponse).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		// SetError(&err).
		Put(createDatabaseURL)

	if err != nil {
		log.Println(err)
		return false
	}

	if resp.IsError() {
		log.Println(resp.ToString())
		return false
	}

	return couchDBResponse.Ok

}

func (ds DatabaseStore[T]) DocumentCreate(key string, document *T) (string, error) {

	documentUrl := fmt.Sprintf("%s/%s/%s", ds.databaseConfig.CouchDBUrl, ds.databaseConfig.DatabaseName, key)

	var couchDBResponse couchdb_client.CouchDBResponse
	//	var err error

	resp, err := ds.httpClient.R().
		SetResult(&couchDBResponse).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		SetBodyJsonMarshal(document).
		// SetError(&err).
		Put(documentUrl)

	if err != nil {
		log.Println(err)
		return "", err
	}

	if resp.IsError() {
		errString, err := resp.ToString()

		if err != nil {
			log.Println(err)
			return "", err
		}

		return "", errors.New(errString)
	}

	return couchDBResponse.Rev, nil
}

func (ds DatabaseStore[T]) DocumentGet(key string) (*T, error) {

	// documentUrl := fmt.Sprintf("%s/%s/%s", ds.couchDBUrl, ds.databaseName, key)
	// log.Println("DocumentGet(", key, ")")
	documentUrl := ds.databaseConfig.DocumentURL(key)

	var responseDocument T

	// var couchDBResponse couchdb_client.CouchDBResponse
	//	var err error

	resp, err := ds.httpClient.R().
		SetResult(&responseDocument).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		// SetError(&err).
		Get(documentUrl)

	if err != nil {
		// log.Println("Error:", err)
		return nil, err
	}

	if resp.IsError() {

		errString, err := resp.ToString()
		// log.Println("Error:", errString)

		if err != nil {
			log.Println("DocumentGet(", key, ") ", err)
			return nil, err
		}

		return nil, errors.New(errString)
	}

	return &responseDocument, nil

}

func (ds DatabaseStore[T]) DocumentUpdate(key string, revision string, document *T) (string, error) {

	documentUrl := ds.databaseConfig.DocumentURL(key)

	var couchDBResponse couchdb_client.CouchDBResponse
	//	var err error

	resp, err := ds.httpClient.R().
		SetResult(&couchDBResponse).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		SetQueryParam("_rev", revision).
		SetBodyJsonMarshal(document).
		Put(documentUrl)

	if err != nil {
		log.Println(err)
		return "", err
	}

	if resp.IsError() {
		errString, err := resp.ToString()

		if err != nil {
			log.Println(err)
			return "", err
		}

		return "", errors.New(errString)
	}

	return couchDBResponse.Rev, nil

}

func (ds DatabaseStore[T]) DocumentDelete(key string, revision string) (string, error) {

	documentUrl := ds.databaseConfig.DocumentURL(key)

	var couchDBResponse couchdb_client.CouchDBResponse
	//	var err error

	resp, err := ds.httpClient.R().
		SetResult(&couchDBResponse).
		SetBasicAuth(ds.databaseConfig.Username, ds.databaseConfig.Password).
		SetQueryParam("rev", revision).
		Delete(documentUrl)

	if err != nil {
		log.Println(err)
		return "", err
	}

	if resp.IsError() {
		errString, err := resp.ToString()

		if err != nil {
			log.Println(err)
			return "", err
		}

		return "", errors.New(errString)
	}

	return couchDBResponse.Rev, nil

}
