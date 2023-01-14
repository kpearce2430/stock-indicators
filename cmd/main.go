package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kpearce2430/keputils/utils"
	"iex-indicators/cmd/internal/handlers/indicators"
	"iex-indicators/cmd/internal/handlers/lookups"
	"iex-indicators/cmd/internal/handlers/portfolio_value"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	router := gin.Default()
	router.GET("/rsi", indicators.GetRsiRouter)
	router.GET("/macd", indicators.GetMACDRouter)
	router.POST("/lookups/:id", lookups.LoadLookups)
	router.GET("/lookups/:id", lookups.GetLookups)
	router.POST("/pv", portfolio_value.LoadPortfolioValueHandler)
	router.GET("/pv/:symbol", portfolio_value.GetPortfolioValueHandler)

	myPort := fmt.Sprintf(":%s", utils.GetEnv("PORT", "8080"))

	srv := &http.Server{
		Addr:    myPort,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
