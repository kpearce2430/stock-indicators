package stock_ind_router

type StatusObject struct {
	Status string `json:"status"`
	Symbol string `json:"symbol,omitempty"`
}
