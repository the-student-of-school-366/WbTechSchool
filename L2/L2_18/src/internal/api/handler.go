package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sdrapkin/guid"
)

func (u *CalenderServer) parseEvent(c *gin.Context) (Event, error) {
	var input struct {
		Date    string `json:"date"`
		Text    string `json:"text"`
		EventId string `json:"event_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		return Event{}, err
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return Event{}, err
	}

	var event Event
	event.Date = date
	event.Text = input.Text

	if input.EventId != "" {
		parsedId, err := guid.Parse(input.EventId)
		if err == nil {
			event.EventId = parsedId
		}
	} else {
		event.EventId = guid.New()
	}
	return event, nil
}

func (u *CalenderServer) CreateEventHandler(c *gin.Context) {
	event, err := u.parseEvent(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = u.SaveEvent(event, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event created": event})
}

func (u *CalenderServer) UpdateEventHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := u.parseEvent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = u.UpdateEvent(userId, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"event updated": event})
	return
}

func (u *CalenderServer) DeleteEventHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	eventId, err := guid.Parse(c.Query("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = u.DeleteEvent(userId, eventId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted event with id": eventId})
}

func (u *CalenderServer) GetEventsForDayHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := u.parseEvent(c)
	date := event.Date
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	events, err := u.GetEventsForDay(userId, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (u *CalenderServer) GetEventsForWeekHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := u.parseEvent(c)
	date := event.Date
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	events, err := u.GetEventsForWeek(userId, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (u *CalenderServer) GetEventsForMonthHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := u.parseEvent(c)
	date := event.Date
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	events, err := u.GetEventsForMonth(userId, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (u *CalenderServer) CreateUserHandler(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u.AddUser(userId)
	c.JSON(http.StatusOK, gin.H{"user created": userId})
}
