package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	couch_database "iex-indicators/couch-database"
	"iex-indicators/lookups"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func LoadLookups(c *gin.Context) {

	rawData, err := ioutil.ReadAll(c.Request.Body)

	id := c.Param("id")

	log.Println("id:", id)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	if err != nil {
		status := StatusObject{Status: "Invalid Lookups Received"}
		log.Println(status)
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	// convert it to a string
	stringData := string(rawData)

	// log.Println(stringData)
	rawLookups := strings.Split(stringData, "\n")
	log.Println("Num Expected Lookups:", len(rawLookups))

	lookupSet := lookups.NewLoadLookupSet(id)

	for i := 0; i < len(rawLookups); i++ {
		lookupRow := rawLookups[i]

		if strings.Contains(lookupRow, "\"") {
			log.Println("Found \" in ", lookupRow)
			idx1 := strings.Index(lookupRow, "\"")
			idx2 := strings.Index(lookupRow[idx1+1:], "\"")
			name := lookupRow[idx1+1 : idx2+1]
			log.Println("name:", name)
			symbol := lookupRow[idx2+3:]
			log.Println("symbol:", symbol)
			lookup := lookups.LoopUpItem{Name: strings.TrimSpace(name), Symbol: strings.TrimSpace(symbol)}
			lookupSet.LookUps = append(lookupSet.LookUps, lookup)
		} else {
			lookupData := strings.Split(lookupRow, ",")
			// log.Println(lookupData[0], ":", lookupData[1])
			lookup := lookups.LoopUpItem{Name: strings.TrimSpace(lookupData[0]), Symbol: strings.TrimSpace(lookupData[1])}
			lookupSet.LookUps = append(lookupSet.LookUps, lookup)
		}
	}

	log.Println("Num Actual lookups>>", len(lookupSet.LookUps))

	var dbConfig couch_database.DatabaseConfig
	err = envconfig.Process("", &dbConfig)
	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	log.Println("Config>>", dbConfig.Username, dbConfig.CouchDBUrl)

	dbConfig.DatabaseName = databaseName

	lookupDatabase := couch_database.NewDataStore[lookups.LookUpSet](&dbConfig)

	_, err = lookupDatabase.DatabaseExists()

	if err != nil {
		if lookupDatabase.DatabaseCreate() == false {

			c.IndentedJSON(http.StatusInternalServerError, "Backend DB Issue")
			return
		}
		log.Println("Database Created")
	} else {
		log.Println("Database Exists")
	}

	lookupRecord, err := lookupDatabase.DocumentGet(id)

	if err != nil {

		log.Println("Creating Document")
		_, err := lookupDatabase.DocumentCreate(id, lookupSet)
		if err != nil {
			// log.Println("Error Creating Document")
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}

	} else {
		lookupSet.Rev = lookupRecord.Rev
		dt := time.Now()
		lookupSet.Timestamp = dt.Format("2006-01-02 15:04:05")
		// log.Println("Updating Document")
		_, err := lookupDatabase.DocumentUpdate(id, lookupRecord.Rev, lookupSet)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.IndentedJSON(http.StatusOK, StatusObject{Status: fmt.Sprintf("ok: %d loaded for %s", len(lookupSet.LookUps), id)})

}

func GetLookups(c *gin.Context) {

	id := c.Param("id")
	log.Println("id:", id)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
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

	lookupDatabase := couch_database.NewDataStore[lookups.LookUpSet](&dbConfig)

	_, err = lookupDatabase.DatabaseExists()

	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	lookupResult, err := lookupDatabase.DocumentGet(id)

	if err != nil {
		log.Println("err>", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	// log.Println("Results:>>", lookupResult)
	c.IndentedJSON(http.StatusOK, lookupResult)

}
