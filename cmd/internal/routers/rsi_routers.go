package routers

import (
	"github.com/gin-gonic/gin"
)

func GetRsiRouter(c *gin.Context) {

	BasicRouter(c, "rsi")

}
