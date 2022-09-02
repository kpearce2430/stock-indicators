package iex_client

//
type IexChartResponse struct {
	Close                float64 `json:"close"`
	High                 float64 `json:"high"`
	Low                  float64 `json:"low"`
	Open                 float64 `json:"open"`
	Symbol               string  `json:"symbol"`
	Volume               int     `json:"volume"`
	Id                   string  `json:"id"`
	Key                  string  `json:"key"`
	Subkey               string  `json:"subkey"`
	Date                 string  `json:"date"`
	Updated              int64   `json:"updated"`
	ChangeOverTime       float64 `json:"changeOverTime"`
	MarketChangeOverTime float64 `json:"marketChangeOverTime"`
	UOpen                float64 `json:"uOpen"`
	UClose               float64 `json:"uClose"`
	UHigh                float64 `json:"uHigh"`
	ULow                 float64 `json:"uLow"`
	UVolume              int     `json:"uVolume"`
	FOpen                float64 `json:"fOpen"`
	FClose               float64 `json:"fClose"`
	FHigh                float64 `json:"fHigh"`
	FLow                 float64 `json:"fLow"`
	FVolume              int     `json:"fVolume"`
	Label                string  `json:"label"`
	Change               float64 `json:"change"`
	ChangePercent        float64 `json:"changePercent"`
}

type IexIndicatorResponse struct {
	Indicator [][]float64        `json:"indicator"`
	Chart     []IexChartResponse `json:"chart"`
}
