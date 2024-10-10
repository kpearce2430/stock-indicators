package model

import (
	"strings"
	"time"
)

type Events struct {
	Date        time.Time
	EventType   string
	FromAccount string
	ToAccount   string
}

func (e *Events) IsFromAccount(tm time.Time, account string) (bool, string) {
	if tm.Year() == e.Date.Year() && tm.Month() == e.Date.Month() && tm.Day() == e.Date.Day() {
		if strings.Compare(account, e.FromAccount) == 0 {
			return true, e.ToAccount
		}
	}
	return false, ""
}

func (e *Events) IsToAccount(tm time.Time, account string) (bool, string) {
	if tm.Year() == e.Date.Year() && tm.Month() == e.Date.Month() && tm.Day() == e.Date.Day() {
		if strings.Compare(account, e.ToAccount) == 0 {
			return true, e.FromAccount
		}
	}
	return false, ""
}
