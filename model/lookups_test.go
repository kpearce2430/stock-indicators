package model_test

import (
	_ "embed"
	"github.com/kpearce2430/stock-tools/model"
	"testing"
)

//go:embed testdata/lookups.csv
var csvLookupData []byte

func TestLoadLookupSet(t *testing.T) {
	t.Parallel()
	ls := model.LoadLookupSet("1", string(csvLookupData))
	if len(ls.LookUps) != 14 {
		t.Log("LookUp Count ", len(ls.LookUps), " does not equal 9")
		t.Fail()
	}
}
