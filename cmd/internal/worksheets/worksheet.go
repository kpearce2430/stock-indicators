package worksheets

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/massive-com/client-go/v2/rest/models"
)

type WorkSheet struct {
	PGXConn    *pgxpool.Pool
	Lookups    *model.LookUpSet
	StockCache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	StockFile  *stocksheet.StockFile
}

func New(conn *pgxpool.Pool) *WorkSheet {
	s := stocksheet.New()
	return &WorkSheet{
		PGXConn:   conn,
		StockFile: s,
	}
}
