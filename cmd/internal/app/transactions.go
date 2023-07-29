package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"iex-indicators/model"
	"io"
	"net/http"
)

func (a *App) LoadTransactionsHandler(c *gin.Context) {

	if a.LookupSet == nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "Lookup Not Loaded"})
		return
	}

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

	// go func() {
	t := model.NewTransactionSet()
	if err = t.Load(rawData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	for _, tr := range t.TransactionRows {
		symbol, ok := a.LookupSet.GetLookUpByName(tr.Security)
		if ok {
			switch {
			case symbol == "DEAD" || symbol == "Missing" || symbol == "Symbol":
				continue
			default:
				tr.Symbol = symbol
			}
		}

		// (t == 'Payment/Deposit' or t == 'Interest Income' or 'Miscellaneous Income'):
		if tr.Type == "Payment/Deposit" || tr.Type == "Interest Income" || tr.Type == "Miscellaneous Income" {
			logrus.Debug("Skipping ", tr.Type)
			continue
		}

		if tr.Symbol == "" {
			logrus.Error("Invalid Entry Symbol:", tr)
			continue
		}

		if tr.Account == "" {
			logrus.Error("Missing Account", tr)
			continue
		}

		ticker, ok := a.Tickers[tr.Symbol]
		if !ok {
			ticker = model.NewTicker(tr.Symbol)
			a.Tickers[tr.Symbol] = ticker
			// logrus.Info("Adding ", e.Symbol)
		}
		ticker.AddEntity(model.NewEntityFromTransaction(tr))
	}

	// }()
	logrus.Info("Number of Tickers> ", len(a.Tickers))
	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: fmt.Sprintf("%d", len(t.TransactionRows))})
}
