package model

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
