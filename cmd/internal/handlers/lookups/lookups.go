package lookups

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"iex-indicators/cmd/internal/stock_ind_router"
	couch_database "iex-indicators/couch-database"
	"iex-indicators/lookups"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func LoadLookups(c *gin.Context) {

	log.Println("In LoadLookups")
	rawData, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		status := stock_ind_router.StatusObject{Status: "Invalid Lookups Received"}
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	id := c.Param("id")

	log.Println("id:", id)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	log.Println("id:", id, " database:", databaseName)

	r := csv.NewReader(strings.NewReader(string(rawData)))
	// This sets the reader to not base the number of fields off the first record.
	r.FieldsPerRecord = -1
	lookupSet := lookups.NewLoadLookupSet(id)

	for {
		record, _ := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if len(record) < 2 {
			// continue
			break
		}

		log.Println(record)

		lookup := lookups.LoopUpItem{Name: strings.TrimSpace(record[0]), Symbol: strings.TrimSpace(record[1])}
		lookupSet.LookUps = append(lookupSet.LookUps, lookup)
	}

	log.Println("Num Actual lookups>>", len(lookupSet.LookUps))

	lookupDatabase, err := couch_database.GetDataStoreByDatabaseName[lookups.LookUpSet](databaseName)

	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

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

	c.IndentedJSON(http.StatusOK, stock_ind_router.StatusObject{Status: fmt.Sprintf("ok: %d loaded for %s", len(lookupSet.LookUps), id)})

}

func GetLookups(c *gin.Context) {

	id := c.Param("id")
	log.Println("id:", id)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	lookupResult, err := GetLookupsFromDatabase(databaseName, id)

	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	c.IndentedJSON(http.StatusOK, lookupResult)

}

func GetLookupName(c *gin.Context) {

	id := c.Param("id")
	log.Println("id:", id)

	name := c.Param("name")
	log.Println("name:", name)

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	log.Println("id:", id, " name:", name, " database:", databaseName)

	lookupResult, err := GetLookupsFromDatabase(databaseName, id)

	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Document Store Issue")
		return
	}

	lookups := lookupResult.LookUps

	// fmt.Println(len(lookups))

	for _, v := range lookups {

		// fmt.Println(i, ":", v)
		if v.Name == name {
			c.IndentedJSON(http.StatusOK, stock_ind_router.StatusObject{Status: "ok", Symbol: v.Symbol})
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, stock_ind_router.StatusObject{Status: "Not Found", Symbol: ""})

}

func GetLookupsFromDatabase(databaseName string, id string) (*lookups.LookUpSet, error) {

	lookupDatabase, err := couch_database.GetDataStoreByDatabaseName[lookups.LookUpSet](databaseName)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	_, err = lookupDatabase.DatabaseExists()

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return lookupDatabase.DocumentGet(id)

}
