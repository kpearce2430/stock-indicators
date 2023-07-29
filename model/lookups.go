package model

import (
	"encoding/csv"
	"io"
	"log"
	"strings"
	"time"
)

type LookUpSet struct {
	Id        string            `json:"_id"`
	Rev       string            `json:"_rev,omitempty"`
	Timestamp string            `json:"timestamp"`
	LookUps   map[string]string `json:"look_ups"`
}

func NewLookupSet(id string) *LookUpSet {

	dt := time.Now()
	curTimeStr := dt.Format("2006-01-02 15:04:05")
	lookupSet := LookUpSet{Id: id, Timestamp: curTimeStr}
	lookupSet.LookUps = make(map[string]string)
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

		// lookup := LoopUpItem{Name: strings.TrimSpace(record[0]), Symbol: strings.TrimSpace(record[1])}
		// lookupSet.LookUps = append(lookupSet.LookUps, lookup)
		lookupSet.LookUps[record[0]] = record[1]
	}
	return lookupSet
}

func (l *LookUpSet) GetLookUpByName(name string) (string, bool) {
	if l.LookUps == nil {
		panic("Missing lookups")
	}
	val, ok := l.LookUps[name]
	return val, ok
}

func (l *LookUpSet) GetLookUpBySymbol(symbol string) (string, bool) {
	if l.LookUps == nil {
		panic("Missing lookups")
	}
	for k, v := range l.LookUps {
		if v == symbol {
			return k, true
		}
	}
	return "", false
}
