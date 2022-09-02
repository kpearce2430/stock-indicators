package couch_database

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"iex-indicators/utils"
)

type DatabaseConfig struct {
	DatabaseName string `envconfig:"DATABASE_NAME",required:"true"`
	CouchDBUrl   string `envconfig:"COUCHDB_URL",default:"http://localhost:5984"`
	Username     string `envconfig:"COUCHDB_USER",default:"admin"`
	Password     string `envconfig:"COUCHDB_PASSWORD",default:"password"`
}

func NewDatabaseConfig(prefix string) (*DatabaseConfig, error) {

	myprefix := prefix

	// One last change to look for a prefix
	if prefix == "" {
		myprefix = utils.GetEnv("PREFIX", "")
	}

	var dbConfig DatabaseConfig
	err := envconfig.Process(myprefix, &dbConfig)

	return &dbConfig, err

}

func (dc DatabaseConfig) DocumentURL(key string) string {

	return fmt.Sprintf("%s/%s/%s", dc.CouchDBUrl, dc.DatabaseName, key)
}
