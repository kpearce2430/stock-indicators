package symbollist

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kpearce2430/stock-tools/model"
	"github.com/sirupsen/logrus"
	"net/http"
)

type SymbolList struct {
	PGXConn *pgxpool.Pool
	Lookups *model.LookUpSet
}

func NewSymbolList(pgxConn *pgxpool.Pool, lookups *model.LookUpSet) *SymbolList {
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

	symbolSet, err := model.SymbolList(c.Request.Context(), s.PGXConn, s.Lookups)
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
	accountList, err := model.AccountList(c.Request.Context(), s.PGXConn)
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

	acctInfo, err := model.AccountInfoGet(c.Request.Context(), s.PGXConn, acctSymbol)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, acctInfo)
}
