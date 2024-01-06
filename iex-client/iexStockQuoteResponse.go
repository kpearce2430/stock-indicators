package iex_client

type IexStockQuoteResponse struct {
	AvgTotalVolume         uint64  `json:"avgTotalVolume,omitempty"`
	CalculationPrice       string  `json:"calculationPrice,omitempty"`
	Change                 float64 `json:"change,omitempty"`
	ChangePercent          float64 `json:"changePercent,omitempty"`
	Close                  float64 `json:"close,omitempty"`
	CloseSource            string  `json:"closeSource,omitempty"`
	CloseTime              uint64  `json:"closeTime,omitempty"`
	CompanyName            string  `json:"companyName,omitempty"`
	Currency               string  `json:"currency,omitempty"`
	DelayedPrice           float64 `json:"delayedPrice,omitempty"`
	DelayedPriceTime       uint64  `json:"delayedPriceTime,omitempty"`
	ExtendedChange         float64 `json:"extendedChange,omitempty"`
	ExtendedChangePercent  float64 `json:"extendedChangePercent,omitempty"`
	ExtendedPrice          float64 `json:"extendedPrice,omitempty"`
	ExtendedPriceTime      uint64  `json:"extendedPriceTime,omitempty"`
	High                   float64 `json:"high,omitempty"`
	HighSource             string  `json:"highSource,omitempty"`
	HighTime               uint64  `json:"highTime,omitempty"`
	IexAskPrice            float64 `json:"iexAskPrice,omitempty"`
	IexAskSize             float64 `json:"iexAskSize,omitempty"`
	IexBidPrice            float64 `json:"iexBidPrice,omitempty"`
	IexBidSize             float64 `json:"iexBidSize,omitempty"`
	IexClose               float64 `json:"iexClose,omitempty"`
	IexCloseTime           uint64  `json:"iexCloseTime,omitempty"`
	IexLastUpdated         uint64  `json:"iexLastUpdated,omitempty"`
	IexMarketPercent       float64 `json:"iexMarketPercent,omitempty"`
	IexOpen                float64 `json:"iexOpen,omitempty"`
	IexOpenTime            uint64  `json:"iexOpenTime,omitempty"`
	IexRealtimePrice       float64 `json:"iexRealtimePrice,omitempty"`
	IexRealtimeSize        int64   `json:"iexRealtimeSize,omitempty"`
	IexVolume              int64   `json:"iexVolume,omitempty"`
	LastTradeTime          uint64  `json:"lastTradeTime,omitempty"`
	LatestPrice            float64 `json:"latestPrice,omitempty"`
	LatestSource           string  `json:"latestSource,omitempty"`
	LatestTime             string  `json:"latestTime,omitempty"`
	LatestUpdate           int64   `json:"latestUpdate,omitempty"`
	LatestVolume           int64   `json:"latestVolume,omitempty"`
	Low                    float64 `json:"low,omitempty"`
	LowSource              string  `json:"lowSource,omitempty"`
	LowTime                uint64  `json:"lowTime,omitempty"`
	MarketCap              uint64  `json:"marketCap,omitempty"`
	OddLotDelayedPrice     float64 `json:"oddLotDelayedPrice,omitempty"`
	OddLotDelayedPriceTime int64   `json:"oddLotDelayedPriceTime,omitempty"`
	Open                   float64 `json:"open,omitempty"`
	OpenTime               uint64  `json:"openTime,omitempty"`
	OpenSource             string  `json:"openSource,omitempty"`
	PeRatio                float64 `json:"peRatio,omitempty"`
	PreviousClose          float64 `json:"previousClose,omitempty"`
	PreviousVolume         uint64  `json:"previousVolume,omitempty"`
	PrimaryExchange        string  `json:"primaryExchange,omitempty"`
	Symbol                 string  `json:"symbol,omitempty"`
	Volume                 int     `json:"volume,omitempty"`
	Week52High             float64 `json:"week52High,omitempty"`
	Week52Low              float64 `json:"week52Low,omitempty"`
	YtdChange              float64 `json:"ytdChange,omitempty"`
	IsUSMarketOpen         bool    `json:"isUSMarketOpen,omitempty"`
}

func (r IexStockQuoteResponse) GetSymbol() string {
	return r.Symbol
}
