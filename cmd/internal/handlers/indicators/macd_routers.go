package indicators

import (
	"github.com/gin-gonic/gin"
)

func GetMACDRouter(c *gin.Context) {

	BasicRouter(c, "macd")

}
