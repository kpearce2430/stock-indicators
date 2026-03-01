package styles_test

import (
	"github.com/kpearce2430/stock-tools/stocksheet/styles"
	"github.com/xuri/excelize/v2"
	"testing"
)

func TestStyles_New(t *testing.T) {
	s, err := styles.New(excelize.NewFile())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(s)
}
