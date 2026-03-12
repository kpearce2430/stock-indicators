package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	business_days "github.com/kpearce2430/keputils/business-days"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets"
	"github.com/kpearce2430/stock-tools/cmd/internal/worksheets/symbol_detail"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
)

func (a *App) CreateSymbolDetailHandler(c *gin.Context) {
	//
	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	worksheetName := c.DefaultQuery("name", "Symbol Detail")
	symbolsList := c.DefaultQuery("symbols", "")
	tableName := c.DefaultQuery("table", "fund_history")

	if symbolsList == "" {
		c.IndentedJSON(http.StatusBadRequest, model.StatusObject{Status: "symbol required"})
		return
	}

	currDay := business_days.GetBusinessDay(time.Now())

	ws := worksheets.New(a.PGXConn)
	ws.Lookups = a.LookupSet
	ws.StockCache = a.StockCache
	// ws.DividendCache = a.DividendCache

	symbols := strings.Split(symbolsList, ",")

	sd := symbol_detail.New(ws)
	for _, s := range symbols {
		if err := sd.SymbolsDetails(fmt.Sprintf("%s %s", s, worksheetName), s, tableName, currDay, 36); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
			return
		}
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
