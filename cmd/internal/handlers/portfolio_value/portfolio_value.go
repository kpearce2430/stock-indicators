package portfolio_value

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	lookup_handlers "iex-indicators/cmd/internal/handlers/lookups"
	"iex-indicators/cmd/internal/stock_ind_router"
	couch_database "iex-indicators/couch-database"
	pv "iex-indicators/portfolio_value"
	"io"
	"io/ioutil"
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

type PortfolioValueDatabaseRecord struct {
	Id         string                   `json:"_id"`
	Rev        string                   `json:"_rev,omitempty"`
	PV         *pv.PortfolioValueRecord `json:"portfolio_value,omitempty"`
	Key        string                   `json:"key"`
	Symbol     string                   `json:"symbol"`
	Julian     string                   `json:"julian"`
	IEXHistory string                   `json:"iex_history,omitempty"`
}

// asciiString returns a string of only ascii characters.
func asciiString(str string) string {

	byteString := []byte(str)
	newByte := []byte("")
	for i := 0; i < len(byteString); i++ {

		if byteString[i] >= 32 && byteString[i] <= 127 {
			newByte = append(newByte, byteString[i])
		}
	}

	return string(newByte)
}

// LoadPortfolioValueHandler handler for loading Portfolio Value exported from Quicken
func LoadPortfolioValueHandler(c *gin.Context) {

	log.Println("In LoadPortfolioValueHandler ")

	rawData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		status := stock_ind_router.StatusObject{Status: "Invalid Lookups Received"}
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	// TODO Make configurable
	lookups, err := lookup_handlers.GetLookupsFromDatabase("lookups", "1")

	lookupMap := make(map[string]string)
	for _, v := range lookups.LookUps {
		lookupMap[v.Name] = v.Symbol
	}

	// Get Query Parameters
	params := c.Request.URL.Query()
	databaseName := params.Get("database")
	if databaseName == "" {
		databaseName = "portfolio_value"
	}

	julDate := params.Get("juldate")

	if julDate == "" {
		now := time.Now()
		julDate = now.Format("2006002")
	}

	log.Println("julDate:,", julDate, " dbName:", databaseName)

	pvDatabase, err := couch_database.GetDataStoreByDatabaseName[PortfolioValueDatabaseRecord](databaseName)
	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	_, err = pvDatabase.DatabaseExists()

	if err != nil {
		if pvDatabase.DatabaseCreate() == false {

			c.IndentedJSON(http.StatusInternalServerError, "Backend DB Issue")
			return
		}
		log.Println("Database Created")
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

			if strings.HasPrefix(record[0], "Price and Holdings") {

				// if julDate == "" {
				log.Println("found:", record[0])
				parts := strings.Split(record[0], ":")
				log.Println(parts[1])
				// }
				continue
			}

			if len(record) > numRows && record[1] == "Symbol" {
				record[0] = "Name"
				numRows = len(record)
				foundHeader = true
				headers = record
			}

		} else {

			record[0] = asciiString(record[0])

			if strings.Compare(record[0], "Cash") == 0 || strings.Compare(record[0], "Totals") == 0 {
				log.Printf("%d Skipping[%s]", len(record[0]), []byte(record[0]))
				continue
			}

			if len(record) == numRows {

				pvRec, err := pv.NewPortfolioValue(headers, record)
				if err != nil {
					log.Println("Error>>", err.Error())
					c.IndentedJSON(http.StatusInternalServerError, stock_ind_router.StatusObject{Status: err.Error()})
				}

				if pvRec.Symbol == "" {
					symbol := lookupMap[pvRec.Name]
					switch symbol {
					case "DEAD":
						log.Println("Found Dead")
						continue
					case "":
						log.Println("Error Missing Symbol for \"", pvRec, "\"")
						continue
					default:
						pvRec.Symbol = symbol
					}

				}

				key := pvRec.Symbol + ":" + julDate

				rec := PortfolioValueDatabaseRecord{Id: key, Key: key, Julian: julDate, Symbol: pvRec.Symbol, PV: pvRec}

				existing, err := pvDatabase.DocumentGet(key)

				switch {
				case err != nil:
					_, err = pvDatabase.DocumentCreate(key, &rec)
				default:
					rec.Rev = existing.Rev
					_, err = pvDatabase.DocumentUpdate(key, existing.Rev, &rec)
				}

				if err != nil {
					log.Println(key, " Error>>", err.Error())
					c.IndentedJSON(http.StatusInternalServerError, stock_ind_router.StatusObject{Status: err.Error()})
				}
			}
		}
	}

	if foundHeader == true {
		c.IndentedJSON(http.StatusOK, stock_ind_router.StatusObject{Status: "ok"})
	} else {
		c.IndentedJSON(http.StatusInternalServerError, stock_ind_router.StatusObject{Status: "hmmm"})
	}

}

func GetPortfolioValueHandler(c *gin.Context) {

	symbol := c.Param("symbol")
	log.Println("symbol:", symbol)

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
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	// log.Println("Config>>", dbConfig.Username, dbConfig.CouchDBUrl)

	dbConfig.DatabaseName = databaseName

	pvDatabase := couch_database.NewDataStore[PortfolioValueDatabaseRecord](&dbConfig)

	_, err = pvDatabase.DatabaseExists()

	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	key := symbol + ":" + julDate

	pvData, err := pvDatabase.DocumentGet(key)

	if err != nil {
		log.Println("Get>>", err.Error())
		errString := fmt.Sprintf("%s", err.Error())
		if strings.Contains(errString, "missing") {
			c.IndentedJSON(http.StatusNotFound, "Not Found")
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	// log.Println(pvData)
	c.IndentedJSON(http.StatusOK, pvData)

}
