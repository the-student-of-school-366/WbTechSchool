package server_tests

import (
	"L2_18/internal/api"
	"testing"
	"time"
)

func setServer() *api.CalenderServer {
	server := api.NewCalenderServer()
	server.AddUser(1)
	server.AddUser(2)
	server.AddUser(3)
	return server
}
func TestAddUser(t *testing.T) {
	server := api.NewCalenderServer()
	userId := 1
	server.AddUser(userId)
	if _, ok := server.UserCalenders[userId]; !ok {
		t.Error("User not found")
	}
}

func TestSaveEvent(t *testing.T) {
	server := setServer()
	event := api.NewEvent(time.Now(), "test event")
	server.SaveEvent(*event, 1)
	user := server.UserCalenders[1]
	if _, ok := user.Events[event.EventId]; !ok {
		t.Error("Event not found")
	}
}

func TestGetEventsForWeek(t *testing.T) {
	server := setServer()

	event1 := api.NewEvent(time.Now().AddDate(0, 0, 1), "test event1")
	event2 := api.NewEvent(event1.Date.AddDate(0, 0, 3), "test event2")
	event3 := api.NewEvent(event1.Date.AddDate(0, 0, 5), "test event3")
	event4 := api.NewEvent(event1.Date.AddDate(0, 0, 6), "test event4")
	event5 := api.NewEvent(event1.Date.AddDate(0, 0, 10), "test event5")

	server.SaveEvent(*event1, 1)
	server.SaveEvent(*event2, 1)
	server.SaveEvent(*event3, 1)
	server.SaveEvent(*event4, 1)
	server.SaveEvent(*event5, 1)

	events, err := server.GetEventsForWeek(1, time.Now())
	if err != nil {
		t.Error(err)
	}
	if len(events) != 4 {
		t.Errorf("Expected 4 events, found: %v", len(events))
	}

}
