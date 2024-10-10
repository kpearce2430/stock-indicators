package app

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/stock-tools/model"
	polygonclient "github.com/kpearce2430/stock-tools/polygon-client"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (a *App) getDividends(symbol string) (model.DividendsSet, error) {
	client := polygonclient.NewPolygonClient("")

	set, err := client.GetDataSet(symbol)
	if err != nil {
		logrus.Error(err.Error())
		return model.DividendsSet{}, err
	}
	var divs []model.Dividends

	err = json.Unmarshal(set, &divs)
	if err != nil {
		logrus.Error(err.Error())
		return model.DividendsSet{}, err
	}

	ds := model.NewDividendsSet(divs)
	return ds, nil
}

func (a *App) GetDividendsFromDB(c *gin.Context) {
	symbol := c.Param("symbol")
	logrus.Info("symbol:", symbol)

	if symbol == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Symbol")
		return
	}

	var ds model.DividendsSet
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
	//symbolList, err := model.SymbolList(context.Background(), a.PGXConn, a.LookupSet)
	//if err != nil {
	//	c.IndentedJSON(http.StatusInternalServerError, err)
	//	return
	//}
	//
	//var sortedSymbols []string
	//for k, _ := range symbolList {
	//	if k != "" {
	//		sortedSymbols = append(sortedSymbols, k)
	//	}
	//}
	//
	//// Needed for the percentage of portfolio formula
	//sort.Strings(sortedSymbols)

	symbolMap, err := model.PortfolioValueGetTypes(a.PGXConn, PortfolioValueDB)
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
