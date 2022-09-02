package lookups

import "time"

type LoopUpItem struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type LookUpSet struct {
	Id        string       `json:"_id"`
	Rev       string       `json:"_rev,omitempty"`
	Timestamp string       `json:"timestamp"`
	LookUps   []LoopUpItem `json:"look_ups"`
}

func NewLoadLookupSet(id string) *LookUpSet {

	dt := time.Now()
	curTimeStr := dt.Format("2006-01-02 15:04:05")
	lookupSet := LookUpSet{Id: id, Timestamp: curTimeStr}
	return &lookupSet

}
