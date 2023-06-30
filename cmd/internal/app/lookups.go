package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	couch_database "github.com/kpearce2430/keputils/couch-database"
	"github.com/sirupsen/logrus"
	"iex-indicators/lookups"
	"iex-indicators/model"
	"io"
	"net/http"
	"time"
)

func (a *App) LoadLookups(c *gin.Context) {
	logrus.Debug("In LoadLookups")
	rawData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		status := model.StatusObject{Status: "Invalid Lookups Received"}
		c.IndentedJSON(http.StatusBadRequest, status)
		return
	}

	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing ID")
		return
	}

	quaryParams := c.Request.URL.Query()
	databaseName := quaryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	lookupSet := lookups.LoadLookupSet(id, string(rawData))
	lookupDatabase, err := couch_database.GetDataStoreByDatabaseName[lookups.LookUpSet](databaseName)

	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}

	_, err = lookupDatabase.DatabaseExists()
	if err != nil {
		if lookupDatabase.DatabaseCreate() == false {
			c.IndentedJSON(http.StatusInternalServerError, "Backend DB Issue")
			return
		}
		logrus.Debug("Database Created")
	} else {
		logrus.Debug("Database Exists")
	}

	lookupRecord, err := lookupDatabase.DocumentGet(id)
	if err != nil {
		// log.Println("Creating Document")
		_, err := lookupDatabase.DocumentCreate(id, lookupSet)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		lookupSet.Rev = lookupRecord.Rev
		dt := time.Now()
		lookupSet.Timestamp = dt.Format("2006-01-02 15:04:05")
		_, err := lookupDatabase.DocumentUpdate(id, lookupRecord.Rev, lookupSet)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.IndentedJSON(http.StatusOK, model.StatusObject{Status: fmt.Sprintf("ok: %d loaded for %s", len(lookupSet.LookUps), id)})
}

func (a *App) GetLookups(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing ID")
		return
	}

	queryParams := c.Request.URL.Query()
	databaseName := queryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	lookupResult, err := a.getLookupsFromDatabase(databaseName, id)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Backend Issue")
		return
	}
	c.IndentedJSON(http.StatusOK, lookupResult)
}

func (a *App) GetLookupName(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing ID")
		return
	}

	name := c.Param("name")
	if name == "" {
		c.IndentedJSON(http.StatusBadRequest, "Missing Name")
		return
	}

	queryParams := c.Request.URL.Query()
	databaseName := queryParams.Get("database")
	if databaseName == "" {
		databaseName = "lookups"
	}

	//
	lookupResult, err := a.getLookupsFromDatabase(databaseName, id)
	if err != nil {
		logrus.Error(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "Document Store Issue")
		return
	}

	l := lookupResult.GetLookUpByName(name)
	if l != nil {
		c.IndentedJSON(http.StatusOK, model.StatusObject{Status: "ok", Symbol: l.Symbol})
		return
	}
	c.IndentedJSON(http.StatusNotFound, model.StatusObject{Status: "Not Found", Symbol: ""})
}

func (a *App) getLookupsFromDatabase(databaseName string, id string) (*lookups.LookUpSet, error) {
	//
	lookupDatabase, err := couch_database.GetDataStoreByDatabaseName[lookups.LookUpSet](databaseName)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	_, err = lookupDatabase.DatabaseExists()
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	return lookupDatabase.DocumentGet(id)
}
