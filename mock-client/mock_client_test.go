package mock_client_test

import (
	"testing"

	mockclient "github.com/kpearce2430/stock-tools/mock-client"
)

func Test_NewMockClient(t *testing.T) {
	t.Parallel()
	client := mockclient.New()
	if client == nil {
		t.Errorf("Expected non-nil client, got nil")
	}

	tickers := []string{"AAPL", "MSFT", "HD"}
	for _, tckr := range tickers {
		t.Run(tckr, func(t *testing.T) {
			t.Parallel()
			response, err := client.GetData(tckr)
			if err != nil {
				t.Errorf("Expected non-nil response, got nil")
			}
			t.Log(string(response))
		})
	}
}
