package app

import (
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/utils"
	"github.com/sirupsen/logrus"
	"iex-indicators/cmd/internal/handlers/indicators"
	"iex-indicators/lookups"
	"net/http"
)

type App struct {
	Srv       *http.Server
	LookupSet *lookups.LookUpSet
}

func (a *App) routes() {
	router := gin.Default()
	router.GET("/rsi", indicators.GetRsiRouter)
	router.GET("/macd", indicators.GetMACDRouter)
	router.POST("/lookups/:id", a.LoadLookups)
	router.GET("/lookups/:id", a.GetLookups)
	router.POST("/pv", a.LoadPortfolioValueHandler)
	router.GET("/pv/:symbol", a.GetPortfolioValueHandler)
	a.Srv.Handler = router
}

func NewApp(port string) *App {
	a := App{
		Srv: &http.Server{
			Addr: port,
		},
		LookupSet: nil,
	}
	a.routes()
	a.setLogging()
	return &a
}

// SetLogging checks to see if the LOG_LEVEL environmental variable is set to
// override the default.
func (a *App) setLogging() {
	logLevel, err := logrus.ParseLevel(utils.GetEnv("LOG_LEVEL", "info"))
	if err != nil {
		defer logrus.Infof("Unknown log level %s, setting to `info`.", logLevel)
	}
	logrus.SetLevel(logLevel)
	fmtr := &logrus.TextFormatter{
		PadLevelText:     true,
		QuoteEmptyFields: true,
	}
	logrus.SetFormatter(fmtr)
	logrus.SetReportCaller(true)
}
