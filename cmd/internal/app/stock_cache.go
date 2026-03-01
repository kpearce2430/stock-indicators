package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/massive-com/client-go/v2/rest/models"
	"github.com/sirupsen/logrus"
)

func (a *App) GetStockCache(c *gin.Context) {

	symbol := c.Param("symbol")
	logrus.Debug("symbol:", symbol)
	if symbol == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}

	var resp *models.GetDailyOpenCloseAggResponse
	var err error

	queryParams := c.Request.URL.Query()
	date := queryParams.Get("date")
	if date != "" {
		logrus.Info("date:", date)
		resp, err = a.StockCache.GetCache(symbol, date)
	} else {
		resp, err = a.StockCache.GetCache(symbol)
	}

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}
