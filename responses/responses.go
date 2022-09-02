package responses

import iex_client "iex-indicators/iex-client"

type CouchIndicatorResponse struct {
	Id             string `json:"_id"`
	Rev            string `json:"_rev,omitempty"`
	StockIndicator string
	IexIndicator   bool
	Period         string
	StockSymbol    string
	Date           string
	IndicatorData  iex_client.IexIndicatorResponse
}
