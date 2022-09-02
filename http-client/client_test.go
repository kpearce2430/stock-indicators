package http_client_test

import (
	"iex-indicators/http-client"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Do stuff BEFORE the tests!")
	exitVal := m.Run()
	log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}

func TestGetDefaultClient(t *testing.T) {

	client := http_client.GetDefaultClient(10, true)
	resp, err := client.R().Get("https://httpstat.us/200")

	if err != nil {
		t.Fatal(err)
	}

	if resp.Status != "200 OK" {
		t.Logf("%+v\n", resp.Status)
		t.Fatal("Response Not 200")
	}

	// t.Logf("%+v\n", resp)
	resp, err = client.R().Get("https://httpstat.us/404")
	if err != nil {
		t.Fatal(err)
	}

	if resp.IsError() {
		t.Logf("Error: %s", resp.Status)
	} else {
		t.Error("Expecting an 404 Not Found Error")
	}

	if resp.Status != "404 Not Found" {
		t.Logf("%+v\n", resp.Status)
		t.Fatal("Response Not 404 Not Found")
	}

}
