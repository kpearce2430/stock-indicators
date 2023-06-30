package model

type StatusObject struct {
	Status string `json:"status"`
	Symbol string `json:"symbol,omitempty"`
}
