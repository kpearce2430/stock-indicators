package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/account"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/dividend_analysis"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/lookups"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/stock_analysis"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/transactionswks"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
)

func (a *App) CreateWorksheetHandler(c *gin.Context) {
	//
	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	worksheetName := c.DefaultQuery("name", "worksheet")
	currDay := business_days.GetBusinessDay(time.Now())
	julDate := c.DefaultQuery("juldate", utils.JulDateFromTime(currDay))
	logrus.Debug("Worksheet ", worksheetName, " Julian Date is:", julDate)

	ws := worksheets.New(a.PGXConn)
	ws.Lookups = a.LookupSet
	ws.StockCache = a.StockCache
	// ws.DividendCache = a.DividendCache

	sa := stock_analysis.New(ws)
	sa.SetStockCache(a.StockCache)
	err := sa.StockAnalysis("Stock Analysis", julDate)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	div := dividend_analysis.New(ws)
	if err = div.DividendSheets(c.Request.Context(), "Dividend Analysis", time.Now(), 48); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	acct := account.New(ws)

	if err := acct.AccountDividends("Account Dividends", time.Now(), 48); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	tr := transactionswks.New(ws)
	if err = tr.Transactions("Transactions", julDate); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	l := lookups.New(ws)
	if err := l.LookupSheet("Lookups"); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err = ws.StockFile.DeleteSheet("Sheet1"); err != nil {
		logrus.Error(err.Error())
	}

	buff, err := ws.GetExcelizeFile().WriteToBuffer()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", worksheetName))
	c.Data(http.StatusOK, "application/octet-stream", buff.Bytes())
}

func (a *App) CreateDividendsHandler(c *gin.Context) {
	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	worksheetName := c.DefaultQuery("name", "worksheet")
	currDay := business_days.GetBusinessDay(time.Now())
	julDate := c.DefaultQuery("juldate", utils.JulDateFromTime(currDay))
	logrus.Info("Worksheet ", worksheetName, " Julian Date is:", julDate)

	ws := worksheets.New(a.PGXConn)
	ws.Lookups = a.LookupSet
	ws.StockCache = a.StockCache

	div := dividend_analysis.New(ws)

	if err := div.DividendSheets(c.Request.Context(), "Dividend Analysis", time.Now(), 48); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.StockFile.DeleteSheet("Sheet1"); err != nil {
		logrus.Error(err.Error())
	}

	buff, err := ws.GetExcelizeFile().WriteToBuffer()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", worksheetName))
	c.Data(http.StatusOK, "application/octet-stream", buff.Bytes())
}
