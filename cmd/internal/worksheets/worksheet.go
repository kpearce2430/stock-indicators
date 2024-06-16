package worksheets

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type WorkSheet struct {
	PGXConn       *pgxpool.Pool
	Lookups       *model.LookUpSet
	File          *excelize.File
	styles        *Styles
	StockCache    *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	DividendCache *stock_cache.Cache[models.Dividend]
}

func NewWorkSheet(f *excelize.File, conn *pgxpool.Pool) *WorkSheet {
	s, err := DefaultStyles(f)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	return &WorkSheet{
		File:    f,
		PGXConn: conn,
		styles:  s,
	}
}
