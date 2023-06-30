package lookups

import (
	"encoding/csv"
	"io"
	"log"
	"strings"
	"time"
)

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

func NewLookupSet(id string) *LookUpSet {

	dt := time.Now()
	curTimeStr := dt.Format("2006-01-02 15:04:05")
	lookupSet := LookUpSet{Id: id, Timestamp: curTimeStr}
	return &lookupSet
}

func LoadLookupSet(id string, csvData string) *LookUpSet {
	r := csv.NewReader(strings.NewReader(csvData))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1
	lookupSet := NewLookupSet(id)

	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			break
		}

		if len(record) < 2 {
			break
		}

		lookup := LoopUpItem{Name: strings.TrimSpace(record[0]), Symbol: strings.TrimSpace(record[1])}
		lookupSet.LookUps = append(lookupSet.LookUps, lookup)
	}
	return lookupSet
}

func (l *LookUpSet) GetLookUpByName(name string) *LoopUpItem {
	if l.LookUps == nil {
		panic("Missing lookups")
	}
	lookups := l.LookUps
	for _, v := range lookups {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (l *LookUpSet) GetLookUpBySymbol(symbol string) *LoopUpItem {
	if l.LookUps == nil {
		panic("Missing lookups")
	}

	for _, v := range l.LookUps {
		if v.Symbol == symbol {
			return &v
		}
	}
	return nil
}
