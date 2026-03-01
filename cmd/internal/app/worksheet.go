package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
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
	logrus.Info("Worksheet ", worksheetName, " Julian Date is:", julDate)

	ws := worksheets.New(a.PGXConn)
	ws.Lookups = a.LookupSet
	ws.StockCache = a.StockCache
	// ws.DividendCache = a.DividendCache

	if err := ws.StockAnalysis("Stock Analysis", julDate); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.DividendSheets(c.Request.Context(), "Dividend Analysis", time.Now(), 48); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.Transactions("Transactions", julDate); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.LookupSheet("Lookups"); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.StockFile.DeleteSheet("Sheet1"); err != nil {
		logrus.Error(err.Error())
	}

	buff, err := ws.StockFile.GetFile().WriteToBuffer()
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

	if err := ws.DividendSheets(c.Request.Context(), "Dividend Analysis", time.Now(), 48); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.StockFile.DeleteSheet("Sheet1"); err != nil {
		logrus.Error(err.Error())
	}

	buff, err := ws.StockFile.GetFile().WriteToBuffer()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", worksheetName))
	c.Data(http.StatusOK, "application/octet-stream", buff.Bytes())
}
