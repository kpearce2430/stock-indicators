package app

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (a *App) PostgresCheck() (bool, error) {
	var count int
	ctx := context.Background()
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s;", TransactionTable)
	if err := a.PGXConn.QueryRow(ctx, countSql).Scan(&count); err != nil {
		return false, err
	}
	logrus.Debug("Count:", count)
	return true, nil
}

func (a *App) CouchDBCheck() bool {
	return a.StockCache.CouchDBUp() && a.DividendCache.CouchDBUp()
}

func (a *App) Status(c *gin.Context) {

	pgxStatus, err := a.PostgresCheck()
	couchStatus := a.CouchDBCheck()

	switch {
	case err != nil:
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	case pgxStatus && couchStatus == false:
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Backends Not Ready"})
		return
	}

	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "OK"})
}
