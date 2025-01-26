package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"net/http"
	"time"
)

func (a *App) AccountDividends(c *gin.Context) {

	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	today := time.Now()
	septFirst23 := time.Date(2023, 9, 01, 00, 00, 00, 00, time.UTC)

	years := (today.Year() - septFirst23.Year()) * 12
	months := int(today.Month()) - (int(septFirst23.Month()))
	monthsAgo := years + months + 1

	worksheetName := c.DefaultQuery("name", "worksheet")

	ws := worksheets.NewWorkSheet(excelize.NewFile(), a.PGXConn)
	ws.Lookups = a.LookupSet
	ws.StockCache = a.StockCache

	if err := ws.AccountDividends("Account Dividends", time.Now(), monthsAgo); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	if err := ws.File.DeleteSheet("Sheet1"); err != nil {
		logrus.Error(err.Error())
	}

	buff, err := ws.File.WriteToBuffer()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", worksheetName))
	c.Data(http.StatusOK, "application/octet-stream", buff.Bytes())
}
