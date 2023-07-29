package app

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"iex-indicators/model"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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
		logrus.Error("Missing Lookup Set")
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
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
		status := model.StatusObject{Status: "Invalid Lookups Received"}
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	// Get Query Parameters
	params := c.Request.URL.Query()
	databaseName := params.Get("database")
	if databaseName == "" {
		databaseName = "portfolio_value"
	}
	julDate := params.Get("juldate")
	logrus.Debugln("julDate:,", julDate, " dbName:", databaseName)

	pvDatabase, err := couch_database.GetDataStoreByDatabaseName[model.PortfolioValueDatabaseRecord](databaseName)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	_, err = pvDatabase.DatabaseExists()
	if err != nil {
		if pvDatabase.DatabaseCreate() == false {
			c.IndentedJSON(http.StatusInternalServerError, "Backend DB Issue")
			return
		}
		logrus.Info("Database Created")
	}

	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1

	foundHeader := false
	numRows := 2
	var headers []string
	for {
		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		if foundHeader == false {

			if julDate == "" && strings.HasPrefix(record[0], "Price and Holdings as of") {

				// if julDate == "" {
				logrus.Debug("found:", record[0])
				parts := strings.Split(record[0], ":")
				str := strings.TrimSpace(parts[1])
				logrus.Debugf("str[%s]", str)

				date, err := time.Parse("2006-01-02", str)

				if err != nil {
					logrus.Error(err)
					c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
					return
				}

				julDate = date.Format("2006002")
				logrus.Info("PV Julian Date:", julDate)
				continue
			}

			if len(record) > numRows && record[1] == "Symbol" {
				record[0] = "Name"
				numRows = len(record)
				foundHeader = true
				headers = record

				if julDate == "" {
					date := time.Now()
					julDate = date.Format("2006002")
				}
			}

		} else {

			record[0] = utils.AsciiString(record[0])

			if strings.Compare(record[0], "Cash") == 0 || strings.Compare(record[0], "Totals") == 0 {
				continue
			}

			if len(record) == numRows {
				pvRec, err := model.NewPortfolioValue(headers, record)
				if err != nil {
					logrus.Error("Error>>", err.Error())
					c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
				}

				if pvRec.Symbol == "" {
					Symbol, ok := a.LookupSet.GetLookUpByName(pvRec.Name)
					if ok {
						switch Symbol {
						case "DEAD":
							logrus.Debug("Found Dead")
							continue
						case "":
							logrus.Debug("Error Missing Symbol for \"", pvRec, "\"")
							continue
						default:
							pvRec.Symbol = Symbol
						}
					}
				}

				key := pvRec.Symbol + ":" + julDate
				rec := model.PortfolioValueDatabaseRecord{Id: key, Key: key, Julian: julDate, Symbol: pvRec.Symbol, PV: pvRec}
				existing, err := pvDatabase.DocumentGet(key)
				switch {
				case err != nil:
					_, err = pvDatabase.DocumentCreate(key, &rec)
				default:
					rec.Rev = existing.Rev
					_, err = pvDatabase.DocumentUpdate(key, existing.Rev, &rec)
				}

				if err != nil {
					logrus.Error(key, " Error>>", err.Error())
					c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: err.Error()})
				}
			}
		}
	}

	if foundHeader == true {
		c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "ok"})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, model.StatusObject{Status: "hmmm"})
	}

}

func (a *App) GetPortfolioValueHandler(c *gin.Context) {

	symbol := c.Param("symbol")
	logrus.Debug("symbol:", symbol)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "portfolio_value"
	}

	julDate := quaryParams.Get("juldate")

	if julDate == "" {

		now := time.Now()
		julDate = now.Format("2006002")
	}

	var dbConfig couch_database.DatabaseConfig
	err := envconfig.Process("", &dbConfig)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	dbConfig.DatabaseName = databaseName
	pvDatabase := couch_database.NewDataStore[model.PortfolioValueDatabaseRecord](&dbConfig)
	_, err = pvDatabase.DatabaseExists()
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	key := symbol + ":" + julDate
	pvData, err := pvDatabase.DocumentGet(key)
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
