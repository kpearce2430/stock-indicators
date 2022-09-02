package couchdb_client

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
)

type CouchDBSeeds struct {
	Seed []string
}

type CouchDBStatus struct {
	Status string
	Seeds  CouchDBSeeds
}

type CouchDBResponse struct {
	Id  string
	Ok  bool
	Rev string
}

type CouchDBHttpClient struct {
	httpClient http.Client
}

func getDefaultTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 200
	t.MaxConnsPerHost = 200
	t.MaxIdleConnsPerHost = 200

	return t
}

func New(timeout time.Duration, transport *http.Transport) CouchDBHttpClient {

	// TODO:  Allow User to pass their own client in.

	couchDBClient := CouchDBHttpClient{}

	if transport == nil {

		couchDBClient.httpClient = http.Client{
			Timeout:   timeout * time.Second,
			Transport: getDefaultTransport(),
		}

	} else {
		couchDBClient.httpClient = http.Client{
			Timeout:   timeout * time.Second,
			Transport: transport,
		}
	}

	return couchDBClient

}

func CouchDBUp(CouchdbURL string, client *CouchDBHttpClient) bool {

	myUrl := CouchdbURL + "/_up"

	if client == nil {
		newClient := New(30, nil)
		client = &newClient
	}

	body, err := client.CouchDBClient("GET", myUrl, "", "", nil, nil)

	if err != nil {
		fmt.Printf("Client Error: %s\n", err)
		return false
	}

	var responseObject CouchDBStatus

	err = json.Unmarshal(body, &responseObject)

	if err != nil {
		fmt.Printf("JSON Error: %s\n", err)
		return false
	}

	fmt.Println(responseObject.Status)
	if responseObject.Status != "ok" {
		return false
	}
	return true
}

func (cdb CouchDBHttpClient) CouchDBClientValid() bool {

	return true

}

//
// CouchDBClient will only handle the byte level for the input and output data.
// Marshalling will be left to the higher level callers.
//
func (cdb CouchDBHttpClient) CouchDBClient(action string, url string, user string, pswd string, headers map[string]string, data []byte) ([]byte, error) {

	// Some initial setup for the client.
	body := []byte{}
	req, err := http.NewRequest(action, url, bytes.NewBuffer(data))

	if err != nil {
		// fmt.Println("http.NewRequest Error:", err)
		return nil, err
	}

	// If user and pswd is supplied, add it to the BasicAuth
	if user != "" && pswd != "" {
		req.SetBasicAuth(user, pswd)
	}

	// Add in any additional headers here.
	// Note:  CouchDB Documentation in some cases require
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := cdb.httpClient.Do(req)

	if err != nil {
		// fmt.Println("client.Do Error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status, ":", resp.StatusCode)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Println("Error on Status:", resp.Status)
		err := errors.New(resp.Status)
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		// fmt.Println("ioutil.ReadAll error:", err)
		return nil, err
	}

	return body, err

}

func UnmarshallCouchDbResponse(body []byte) (*CouchDBResponse, error) {

	var response CouchDBResponse

	err := json.Unmarshal(body, &response)

	if err != nil {
		log.Printf("Error Unmarshalling: %s", err)
		return nil, err
	}

	return &response, nil
}