package model

import "time"

type Lot struct {
	NumberShares  float64
	PricePerShare float64
	SoldDate      time.Time
}

func (l *Lot) Proceeds() float64 {
	return l.NumberShares * l.PricePerShare
}
