package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	massive_client "github.com/kpearce2430/stock-tools/massive-client"
	"github.com/kpearce2430/stock-tools/model/dividends"
	"github.com/kpearce2430/stock-tools/model/portfolio_value"
	"github.com/sirupsen/logrus"
)

func (a *App) getDividends(symbol string) (dividends.DividendsSet, error) {
	client := massive_client.New()

	set, err := client.GetDataSet(symbol)
	if err != nil {
		logrus.Error(err.Error())
		return dividends.DividendsSet{}, err
	}
	var divs []dividends.Dividends

	err = json.Unmarshal(set, &divs)
	if err != nil {
		logrus.Error(err.Error())
		return dividends.DividendsSet{}, err
	}

	ds := dividends.NewDividendsSet(divs)
	return ds, nil
}

func (a *App) GetDividendsFromDB(c *gin.Context) {
	symbol := c.Param("symbol")
	logrus.Info("symbol:", symbol)

	if symbol == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}

	var ds dividends.DividendsSet
	err := ds.FromDBbySymbol(context.Background(), a.PGXConn, "dividends", symbol)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, ds)
}

func (a *App) GetDividends(c *gin.Context) {
	symbol := c.Param("symbol")
	logrus.Info("symbol:", symbol)

	if symbol == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}

	ds, err := a.getDividends(symbol)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, ds)

	go func() {
		err := ds.ToDB(context.Background(), a.PGXConn, "dividends")
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Info("loaded ", len(ds.Dividends), " for ", symbol)
	}()
}

func (a *App) GetAllDividends(c *gin.Context) {
	symbolMap, err := portfolio_value.GetTypes(a.PGXConn, PortfolioValueDB)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}

	for symbol, v := range symbolMap {
		logrus.Info("symbol:", symbol, " type:", v)

		if v != "Stock" {
			logrus.Info("Skipping:", symbol)
			continue
		}

		ds, err := a.getDividends(symbol)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		//c.IndentedJSON(http.StatusOK, ds)
		//
		//go func() {
		err = ds.ToDB(context.Background(), a.PGXConn, "dividends")
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		logrus.Info("loaded ", len(ds.Dividends), " for ", symbol)
		// break
		//}()
	}
	c.IndentedJSON(http.StatusOK, symbolMap)
}
