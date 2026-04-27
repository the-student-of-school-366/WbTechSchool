package main

import (
	"L2_18/internal/api"
	"L2_18/internal/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	router := gin.Default()
	router.Use(logger.LoggerMiddleware())

	server := api.NewCalenderServer()

	router.POST("/event/:id", server.CreateEventHandler)
	router.PUT("/event/:id", server.UpdateEventHandler)
	router.DELETE("/event/:id", server.DeleteEventHandler)
	router.GET("/event/for_day/:id", server.GetEventsForDayHandler)
	router.GET("/event/for_week/:id", server.GetEventsForWeekHandler)
	router.GET("/event/for_month/:id", server.GetEventsForMonthHandler)

	router.POST("/user/:id", server.CreateUserHandler)
	router.Run(":" + port)
}
