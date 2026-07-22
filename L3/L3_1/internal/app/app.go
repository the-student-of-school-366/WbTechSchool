package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
	server *http.Server
}
