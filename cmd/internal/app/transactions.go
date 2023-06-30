package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"iex-indicators/model"
	"iex-indicators/transactionset"
	"io"
	"net/http"
)

func (a *App) LoadTransactionsHandler(c *gin.Context) {

	defer func() {
		if c != nil && c.Request != nil && c.Request.Body != nil {
			if err := c.Request.Body.Close(); err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// go func() {
	t := transactionset.NewTransactionSet()
	t.Load(rawData)
	// }()

	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: fmt.Sprintf("%d", len(t.TransactionRows))})
}
