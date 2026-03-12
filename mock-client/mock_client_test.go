package mock_client_test

import (
	"testing"

	mockclient "github.com/kpearce2430/stock-tools/mock-client"
)

func Test_NewMockClient(t *testing.T) {
	client := mockclient.New()
	if client == nil {
		t.Errorf("Expected non-nil client, got nil")
	}

	response, err := client.GetData("AAPL")
	if err != nil {
		t.Errorf("Expected non-nil response, got nil")
	}

	t.Log(string(response))
}
