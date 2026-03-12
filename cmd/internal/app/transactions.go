package app

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/kpearce2430/stock-tools/model/transaction"
	"github.com/sirupsen/logrus"
)

func (a *App) LoadTransactionsHandler(c *gin.Context) {
	//
	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

	databaseName := c.DefaultQuery("database", TransactionTable)

	defer func() {
		if c != nil && c.Request != nil && c.Request.Body != nil {
			if err := c.Request.Body.Close(); err != nil {
				logrus.Error(err.Error())
			}
		}
	}()

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ts := transaction.NewTransactionSet()
	if err = ts.LoadToDB(a.PGXConn, a.LookupSet, databaseName, rawData); err != nil {
		e := make(map[string]string)
		e["error"] = err.Error()
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "completed"})
}
