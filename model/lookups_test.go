package model_test

import (
	_ "embed"
	"iex-indicators/model"
	"testing"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

func TestLoadLookupSet(t *testing.T) {

	ls := model.LoadLookupSet("1", string(csvLookupData))

	if len(ls.LookUps) != 9 {
		t.Log("LookUp Count ", len(ls.LookUps), " does not equal 9")
		t.Fail()
	}
}
