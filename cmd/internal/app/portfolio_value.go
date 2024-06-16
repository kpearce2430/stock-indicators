package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/utils"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

var (
	errMissingLookups = fmt.Errorf("missing lookup set")
	errInvalidLookups = fmt.Errorf("invalid lookups received")
)

/*
   "Name": "name",
   "Symbol": "symbol",
   "Type": "type",
   "Price": "quote",
   "Quote": "quote",
   "Price Day Change": "price_day_change",
   "Price Day Change (%)": "price_day_change_pct",
   "Shares": "shares",
   "Cost Basis": "cost_basis",
   "Market Value": "market_value",
   "Average Cost Per Share": "avg_cost_per_share",
   "Gain/Loss 12-Month": "gain_loss_last_12m",
   "Gain/Loss": "gain_loss",
   "Gain/Loss (%)": "gain_loss_pct",
*/

// LoadPortfolioValueHandler handler for loading Portfolio Value exported from Quicken
func (a *App) LoadPortfolioValueHandler(c *gin.Context) {
	logrus.Debug("In LoadPortfolioValueHandler ")
	if a.LookupSet == nil {
		logrus.Error(errMissingLookups.Error())
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: errMissingLookups.Error()})
		return
	}

	defer func() {
		if err := c.Request.Body.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusBadRequest, model.StatusObject{Status: errInvalidLookups.Error()})
		return
	}

	// Get Query Parameters
	databaseName := c.DefaultQuery("database", PortfolioValueDB)
	julDate := c.DefaultQuery("juldate", "")
	logrus.Debugln("julDate:,", julDate, " dbName:", databaseName)

	if err := model.LoadPortfolioValues(databaseName, string(rawData), julDate, a.LookupSet); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "ok"})
}

func (a *App) GetPortfolioValueHandler(c *gin.Context) {
	symbol := c.Param("symbol")
	logrus.Debug("symbol:", symbol)
	c.DefaultQuery("database", PortfolioValueDB)
	julDate := c.DefaultQuery("juldate", utils.JulDate())
	pvData, err := model.GetPortfolioValue(symbol, julDate)

	if err != nil {
		logrus.Debug("Get>>", err.Error())
		errString := fmt.Sprintf("%s", err.Error())
		if strings.Contains(errString, "missing") {
			c.IndentedJSON(http.StatusNotFound, "Not Found")
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}
	c.IndentedJSON(http.StatusOK, pvData)
}

func (a *App) LoadDBPortfolioValueHandler(c *gin.Context) {
	logrus.Debug("In LoadDBPortfolioValueHandler ")
	if a.LookupSet == nil {
		logrus.Error(errMissingLookups.Error())
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: errMissingLookups.Error()})
		return
	}

	databaseName := c.DefaultQuery("database", PortfolioValueDB)
	logrus.Info("database:", databaseName)

	defer func() {
		if err := c.Request.Body.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusBadRequest, model.StatusObject{Status: errInvalidLookups.Error()})
		return
	}

	// Get Query Parameters
	julDate := c.DefaultQuery("juldate", "")
	logrus.Debugln("julDate:,", julDate, " dbName:", databaseName)

	count, err := model.PortfolioValuesLoadDB(a.PGXConn, PortfolioValueDB, string(rawData), julDate, a.LookupSet)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
		return
	}
	logrus.Info("Loaded ", count, " records.")
	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "ok"})
}
