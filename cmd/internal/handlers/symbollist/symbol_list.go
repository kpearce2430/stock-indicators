package symbollist

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/model/account_info"
	"github.com/kpearce2430/stock-tools/model/lookups"
	"github.com/sirupsen/logrus"
)

type SymbolList struct {
	PGXConn *pgxpool.Pool
	Lookups *lookups.LookUpSet
}

func NewSymbolList(pgxConn *pgxpool.Pool, lookups *lookups.LookUpSet) *SymbolList {
	return &SymbolList{
		PGXConn: pgxConn,
		Lookups: lookups,
	}
}

func (s *SymbolList) SymbolListGet(c *gin.Context) {
	if s.Lookups == nil {
		panic("missing lookups")
	}
	if s.PGXConn == nil {
		panic("missing pg connection")
	}

	symbolSet, err := account_info.SymbolList(c.Request.Context(), s.PGXConn, s.Lookups)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, symbolSet)
}

func (s *SymbolList) AccountListGet(c *gin.Context) {
	if s.PGXConn == nil {
		panic("missing pg connection")
	}
	accountList, err := account_info.AccountList(c.Request.Context(), s.PGXConn)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, accountList)
}

func (s *SymbolList) TickerInfoGet(c *gin.Context) {
	acctSymbol := c.Param("symbol")
	logrus.Info("symbol:", acctSymbol)
	// julDate := c.DefaultQuery("juldate", utils.JulDate())

	acctInfo, err := account_info.AccountInfoGet(c.Request.Context(), s.PGXConn, acctSymbol)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, acctInfo)
}
