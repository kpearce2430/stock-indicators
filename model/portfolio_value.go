package model

import "github.com/kpearce2430/keputils/utils"

var PVSymbolMaps = map[string]string{
	"Name":                   "name",
	"Symbol":                 "symbol",
	"Type":                   "type",
	"Price":                  "quote",
	"Quote":                  "quote",
	"Price Day Change":       "price_day_change",
	"Price Day Change (%)":   "price_day_change_pct",
	"Shares":                 "shares",
	"Cost Basis":             "cost_basis",
	"Market Value":           "market_value",
	"Average Cost Per Share": "avg_cost_per_share",
	"Gain/Loss 12-Month":     "gain_loss_last_12m",
	"Gain/Loss":              "gain_loss",
	"Gain/Loss (%)":          "gain_loss_pct",
}

type PortfolioValueRecord struct {
	Name                string  `json:"name"`
	Symbol              string  `json:"symbol"`
	Type                string  `json:"type"`
	Quote               float64 `json:"quote"`
	PriceDayChange      float64 `json:"price_day_change"`
	PriceDayChangePct   float64 `json:"price_day_change_pct"`
	Shares              float64 `json:"shares"`
	CostBasis           float64 `json:"cost_basis"`
	MarketValue         float64 `json:"market_value"`
	AverageCostPerShare float64 `json:"avg_cost_per_share"`
	GainLoss12Month     float64 `json:"gain_loss_last_12m"`
	GainLoss            float64 `json:"gain_loss"`
	GainLossPct         float64 `json:"gain_loss_pct"`
}

type PortfolioValueDatabaseRecord struct {
	Id         string                `json:"_id"`
	Rev        string                `json:"_rev,omitempty"`
	PV         *PortfolioValueRecord `json:"portfolio_value,omitempty"`
	Key        string                `json:"key"`
	Symbol     string                `json:"symbol"`
	Julian     string                `json:"julian"`
	IEXHistory string                `json:"iex_history,omitempty"`
}

// NewPortfolioValue create a new PortfolioValueRecord from the headers and row values provided.
func NewPortfolioValue(headers []string, values []string) (*PortfolioValueRecord, error) {
	pv := PortfolioValueRecord{}
	for index, value := range headers {
		switch value {
		case "Name":
			pv.Name = values[index]
		case "Symbol":
			pv.Symbol = values[index]
		case "Type":
			pv.Type = values[index]
		case "Price":
			pv.Quote, _ = utils.FloatParse(values[index])
		case "Quote":
			pv.Quote, _ = utils.FloatParse(values[index])

		case "Price Day Change":
			pv.PriceDayChange, _ = utils.FloatParse(values[index])

		case "Price Day Change (%)":
			pv.PriceDayChangePct, _ = utils.FloatParse(values[index])

		case "Shares":
			pv.Shares, _ = utils.FloatParse(values[index])

		case "Cost Basis":
			pv.CostBasis, _ = utils.FloatParse(values[index])

		case "Market Value":
			pv.MarketValue, _ = utils.FloatParse(values[index])

		case "Average Cost Per Share":
			pv.AverageCostPerShare, _ = utils.FloatParse(values[index])

		case "Gain/Loss 12-Month":
			pv.GainLoss12Month, _ = utils.FloatParse(values[index])

		case "Gain/Loss":
			pv.GainLoss, _ = utils.FloatParse(values[index])

		case "Gain/Loss (%)":
			pv.GainLossPct, _ = utils.FloatParse(values[index])

		} // switch
	} // for
	return &pv, nil
}
