package model

import "time"

type Lot struct {
	NumberShares  float64
	PricePerShare float64
	SoldDate      time.Time
}
