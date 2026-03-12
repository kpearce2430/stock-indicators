package worksheets

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/kpearce2430/stock-tools/stock_cache"
	"github.com/kpearce2430/stock-tools/stocksheet"
	"github.com/kpearce2430/stock-tools/stocksheet/styles"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/xuri/excelize/v2"
)

type WorkSheet struct {
	PGXConn    *pgxpool.Pool
	Lookups    *lookups.LookUpSet
	StockCache *stock_cache.Cache[models.GetDailyOpenCloseAggResponse]
	StockFile  *stocksheet.StockFile
}

type WorksheetInterface interface {
	GetFile() *stocksheet.StockFile
	GetStyles() *styles.Styles
	GetExcelizeFile() *excelize.File
	GetPGXConn() *pgxpool.Pool
	GetLookups() *lookups.LookUpSet
}

func New(conn *pgxpool.Pool) *WorkSheet {
	s := stocksheet.New()
	return &WorkSheet{
		PGXConn:   conn,
		StockFile: s,
	}
}

func (w *WorkSheet) GetFile() *stocksheet.StockFile {
	return w.StockFile
}

func (w *WorkSheet) GetStyles() *styles.Styles {
	return w.StockFile.GetStyles()
}

func (w *WorkSheet) GetExcelizeFile() *excelize.File {
	return w.StockFile.GetExcelizeFile()
}

func (w *WorkSheet) GetPGXConn() *pgxpool.Pool {
	return w.PGXConn
}

func (w *WorkSheet) GetLookups() *lookups.LookUpSet {
	return w.Lookups
}
