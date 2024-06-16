package app

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (a *App) GetDividendCache(c *gin.Context) {
	symbol := c.Param("symbol")
	logrus.Info("symbol:", symbol)

	if a.DividendCache == nil {
		c.IndentedJSON(http.StatusInternalServerError, "Dividend Cache Not Loaded")
		return
	}
	if symbol == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}

	resp, err := a.DividendCache.GetCacheSet(symbol)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, resp)
}
