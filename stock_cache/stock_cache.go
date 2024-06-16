package stock_cache

import (
	"encoding/json"
	"fmt"
	business_days "github.com/kpearce2430/keputils/business-days"
	couchdatabase "github.com/kpearce2430/keputils/couch-database"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"time"
)

type CacheClient interface {
	GetData(ticker string, args ...string) ([]byte, error)
	GetDataSet(ticker string, args ...string) ([]byte, error)
	GetIndicator(indicator, ticker string, data []byte) ([]byte, error)
}

type Cache[T any] struct {
	couchdatabase.DatabaseStore[T]
	client CacheClient
}

func NewCache[T any](dataConfig *couchdatabase.DatabaseConfig, client CacheClient) (*Cache[T], error) {
	dataStore := couchdatabase.NewDataStore[T](dataConfig)
	return &Cache[T]{
		DatabaseStore: dataStore,
		client:        client,
	}, nil
}

func (c *Cache[T]) GetCache(ticker string, args ...string) (*T, error) {
	logrus.Debug(len(args), ":", args)
	var key string
	// TODO:  Current assumption is that the first argument will be the key.  Figure out a better way in case more args are needed.
	switch len(args) {
	case 0:
		key = fmt.Sprintf("%s:%s", ticker, utils.JulDateFromTime(business_days.GetBusinessDay(time.Now())))
	case 1:
		key = fmt.Sprintf("%s:%s", ticker, args[0])
	default:
		logrus.Error("Invalid arguments:", args)
		return nil, fmt.Errorf("invalid arguments")
	}

	doc, err := c.DocumentGet(key)
	if err != nil {
		logrus.Debug("Error getting document ", key, ":", err.Error())
		return doc, err
	}

	if doc != nil {
		return doc, nil
	}

	resp, err := c.client.GetData(ticker, args...)
	if err != nil {
		logrus.Debug("Error from client.GetStockQuote( ", key, ") :", err.Error())
		return nil, err
	}
	var response T
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	id, err := c.DocumentCreate(key, &response)
	if err != nil {
		return nil, err
	}
	logrus.Debug("Added ", id)
	return &response, nil
}

func (c *Cache[T]) GetCacheSet(ticker string, args ...string) (*T, error) {
	logrus.Debug(len(args), ":", args)
	var key string
	// TODO:  Current assumption is that the first argument will be the key.  Figure out a better way in case more args are needed.
	switch len(args) {
	case 0:
		jDate := utils.JulDateFromTime(business_days.GetBusinessDay(time.Now()))
		key = fmt.Sprintf("%s:%s", ticker, jDate)
	case 1:
		key = fmt.Sprintf("%s:%s", ticker, args[0])
	default:
		logrus.Error("Invalid arguments:", args)
		return nil, fmt.Errorf("invalid arguments")
	}

	doc, err := c.DocumentGet(key)
	if err != nil {
		logrus.Debug("Error getting document ", key, ":", err.Error())
		return doc, err
	}

	if doc != nil {
		return doc, nil
	}

	resp, err := c.client.GetDataSet(ticker)
	if err != nil {
		logrus.Debug("Error from client.GetDataSet( ", key, ") :", err.Error())
		return nil, err
	}
	var responses []T
	if err := json.Unmarshal(resp, &responses); err != nil {
		return nil, err
	}

	if len(responses) > 0 {
		for _, r := range responses {
			newKey := fmt.Sprintf("%s:%s", utils.JulDate(), ticker)
			id, err := c.DocumentCreate(newKey, &r)
			if err != nil {
				return nil, err
			}
			logrus.Debug("Added ", id)
			break //TODO - Handle more than one record
		}
		return &responses[0], nil
	}
	return nil, fmt.Errorf("no records found")
}

/*
func (c *Cache[T]) GetIndicator(args ...string) (*T, error) {

	sym := []string{"symbol", "type", "time"}

	ind := make(map[string]string)

	if len(args) == 0 {
		logrus.Error("invalid number of arguments")
		return nil, fmt.Errorf("invalid number of arguments")
	}

	for i, v := range args {
		ind[sym[i]] = v
	}

	b, err := json.Marshal(ind)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	resp, err := c.client.GetIndicator(b)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}

	var response T
	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

*/
