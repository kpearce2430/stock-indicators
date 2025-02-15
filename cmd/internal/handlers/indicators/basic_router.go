package indicators

import (
	"fmt"
	"github.com/gin-gonic/gin"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	iex_client "github.com/kpearce2430/stock-tools/iex-client"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"strconv"
	"time"
)

// BasicRouter performs the underlying routing functions for any of the IEX stock indicators.
func BasicRouter(c *gin.Context, stockIndicator string) {

	quaryParams := c.Request.URL.Query()
	log.Println(quaryParams)

	symbol := quaryParams.Get("symbol")
	julDate := quaryParams.Get("julDate")
	iexIndicatorOnlyValue := quaryParams.Get("indicatorOnly")

	iexIndicator := false
	var err error

	if iexIndicatorOnlyValue != "" {
		iexIndicator, err = strconv.ParseBool(iexIndicatorOnlyValue)
		if err != nil {
			status := model.StatusObject{Status: "Invalid indicatorOnly"}
			log.Println(status)
			c.IndentedJSON(http.StatusBadRequest, status)
			return
		}
	}

	iexPeriod := quaryParams.Get("period")
	actionIndicator := quaryParams.Get("action")
	// action
	//   none - retrieve current document from couchdb, then iex
	//   update - update current document in couchdb with iex
	//   delete - tbd
	//
	if symbol == "" {
		status := model.StatusObject{Status: "missing symbol"}
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	if julDate == "" {
		julDate = utils.JulDate()
	}

	key := fmt.Sprintf("%s:%s:%s", stockIndicator, symbol, julDate)

	log.Printf("key: %s", key)

	indicatorDatabase := couchdatabase.DataStore[iex_client.CouchIndicatorResponse]("")

	if indicatorDatabase.CouchDBUp() != true {
		// log.Fatal(fmt.Sprintf("%s %s %s %s", couch_database.))
		log.Printf("%s", indicatorDatabase.GetConfig())
		status := model.StatusObject{Status: "CouchDB Not Up"}
		c.IndentedJSON(http.StatusInternalServerError, status)
		return
	}

	_, err = indicatorDatabase.DatabaseExists()
	if err != nil {
		log.Printf("%s", indicatorDatabase.GetConfig())
		// status := stock_ind_router.StatusObject{Status: err.Error()}
		log.Println("Error:", err.Error())
		status := model.StatusObject{Status: "CouchDB Not Up"}
		c.IndentedJSON(http.StatusInternalServerError, status)
		return
	}

	indData, err := indicatorDatabase.DocumentGet(key)

	if err != nil {

		// There is now document found, go off and get one and add it to CouchDB
		log.Printf("Error: %+v", err)
		indData, err = callIexIndicator(stockIndicator, key, symbol, iexIndicator, iexPeriod)

		if err != nil {
			status := model.StatusObject{Status: "IEX Failure"}
			c.IndentedJSON(http.StatusNotFound, status)
			return
		}

		_, err := indicatorDatabase.DocumentCreate(key, indData)

		if err != nil {
			log.Println(err)
			status := model.StatusObject{Status: fmt.Sprintf("%s", err)}
			c.IndentedJSON(http.StatusInternalServerError, status)
		}

		// log.Println(dbResponse)
	} else if actionIndicator == "update" {
		// There is a document found but the caller wants an updated with the based on
		// the latest request.
		if indData == nil {
			status := model.StatusObject{Status: "unexpected nil value from indicator"}
			logrus.Error(status)
			c.IndentedJSON(http.StatusInternalServerError, status)
		}
		logrus.Error("Old Rev:", indData)
		revision := indData.Rev

		indicatorResponse, err := callIexIndicator(stockIndicator, key, symbol, iexIndicator, iexPeriod)

		if err != nil {
			status := model.StatusObject{Status: "IEX Failure"}
			c.IndentedJSON(http.StatusNotFound, status)
			return
		}

		indicatorResponse.Rev = revision

		dbResponse, err := indicatorDatabase.DocumentUpdate(key, indicatorResponse.Rev, indicatorResponse)

		if err != nil {
			log.Println(err)
			status := model.StatusObject{Status: fmt.Sprintf("%s", err)}
			c.IndentedJSON(http.StatusInternalServerError, status)
		}
		log.Printf("CouchDB Response: %+v", dbResponse)

		indicatorResponse, err = indicatorDatabase.DocumentGet(key)

		if err != nil {
			log.Println(err)
			status := model.StatusObject{Status: fmt.Sprintf("%s", err)}
			c.IndentedJSON(http.StatusInternalServerError, status)
		}

		log.Printf("New Rev: %s", indicatorResponse.Rev)
		c.IndentedJSON(http.StatusOK, indicatorResponse)
		return

	}

	// log.Printf(">> %+v", rsiData)
	c.IndentedJSON(http.StatusOK, indData)

}

func callIexIndicator(indicatorType string, key string, symbol string, indicator bool, period string) (*iex_client.CouchIndicatorResponse, error) {

	domain := utils.GetEnv("IEX_URL", "sandbox.iexapis.com")
	iexClient := iex_client.New(domain, 60, true).
		Indicator(indicator).
		Period(period).
		Symbol(symbol)

	log.Println("Indicator Type", indicatorType)
	response, err := iexClient.GetIndicator(indicatorType)

	if err != nil {
		log.Println("Didnt work")
		return nil, err
	}

	indData := iex_client.IexIndicatorResponse{
		Indicator: response.Indicator,
		Chart:     response.Chart,
	}

	dt := time.Now()

	data := iex_client.CouchIndicatorResponse{
		Id:             key,
		StockSymbol:    symbol,
		StockIndicator: indicatorType,
		IexIndicator:   indicator,
		Period:         period,
		Date:           dt.String(),
		IndicatorData:  indData,
	}

	return &data, nil

}
