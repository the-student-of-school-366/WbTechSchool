package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Новый запрос:", c.Request.Method, c.Request.URL.Path, time.Now())
		c.Next()
	}
}
