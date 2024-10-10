package model

import (
	"fmt"
	"strings"
	"time"
)

const dateToPgLayout = "2006-01-02"

type StockTime struct {
	time.Time
}

func (dt *StockTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		dt.Time = time.Time{}
		return
	}
	dt.Time, err = time.Parse(dateToPgLayout, s)
	return
}

func (dt *StockTime) MarshalJSON() ([]byte, error) {
	if dt.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", dt.Format(dateToPgLayout))), nil
}

func (dt *StockTime) TimeConv(t time.Time) {
	dt.Time = t
}
