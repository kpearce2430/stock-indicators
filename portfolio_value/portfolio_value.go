package portfolio_value

import (
	"log"
	"strconv"
	"strings"
)

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

//func cleanUp(s string) string {
//
//	t := strings.Replace(s, ",", "", -1)
//	t = strings.Replace(t, "$", "", -1)
//	t = strings.Replace(t, "#", "", -1)
//	t = strings.Replace(t, "%", "", -1)
//
//	return t
//
//}

// floatParse common float parser
func floatParse(inputString string) (float64, error) {

	t := strings.Replace(inputString, ",", "", -1)
	t = strings.Replace(t, "$", "", -1)
	t = strings.Replace(t, "#", "", -1)
	t = strings.Replace(t, "%", "", -1)

	switch t {
	case "N/A":
		return 0.00, nil

	case "":
		return 0.00, nil

	case "Add":
		return 0.00, nil

	default:
		value, err := strconv.ParseFloat(t, 64)
		if err != nil {
			log.Println("WARNING:", err.Error(), " for[", inputString, "]")
			value = 0.00
		}

		return value, err
	}
}

// NewPortfolioValue create a new PortfolioValueRecord from the headers and row values provided.
func NewPortfolioValue(headers []string, values []string) (*PortfolioValueRecord, error) {

	// log.Println("In NewPortfolioValue", values)

	pv := PortfolioValueRecord{}
	// var err error

	for index, value := range headers {

		switch value {
		case "Name":
			pv.Name = values[index]
		case "Symbol":
			pv.Symbol = values[index]
		case "Type":
			pv.Type = values[index]
		case "Price":
			pv.Quote, _ = floatParse(values[index])
		case "Quote":
			pv.Quote, _ = floatParse(values[index])

		case "Price Day Change":
			pv.PriceDayChange, _ = floatParse(values[index])

		case "Price Day Change (%)":
			pv.PriceDayChangePct, _ = floatParse(values[index])

		case "Shares":
			pv.Shares, _ = floatParse(values[index])

		case "Cost Basis":
			pv.CostBasis, _ = floatParse(values[index])

		case "Market Value":
			pv.MarketValue, _ = floatParse(values[index])

		case "Average Cost Per Share":
			pv.AverageCostPerShare, _ = floatParse(values[index])

		case "Gain/Loss 12-Month":
			pv.GainLoss12Month, _ = floatParse(values[index])

		case "Gain/Loss":
			pv.GainLoss, _ = floatParse(values[index])

		case "Gain/Loss (%)":
			pv.GainLossPct, _ = floatParse(values[index])

		} // switch
	} // for

	return &pv, nil

}
