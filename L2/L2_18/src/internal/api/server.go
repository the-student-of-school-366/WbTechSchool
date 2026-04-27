package api

import (
	"errors"
	"sync"
	"time"

	"github.com/sdrapkin/guid"
)

type CalenderServer struct {
	sync.RWMutex
	UserCalenders map[int]UserCalender
}

func NewCalenderServer() *CalenderServer {
	return &CalenderServer{UserCalenders: make(map[int]UserCalender)}
}

type UserCalender struct {
	UserId int
	Events map[guid.Guid]Event
}

func NewUserCalender(id int) *UserCalender {
	return &UserCalender{
		UserId: id,
		Events: make(map[guid.Guid]Event),
	}
}

type Event struct {
	EventId guid.Guid `json:"event_id"`
	Date    time.Time `json:"date"`
	Text    string    `json:"text"`
}

func NewEvent(date time.Time, text string) *Event {
	return &Event{
		EventId: guid.New(),
		Date:    date,
		Text:    text,
	}
}

func (server *CalenderServer) SaveEvent(event Event, userId int) error {
	server.Lock()
	defer server.Unlock()
	user, ok := server.UserCalenders[userId]
	if !ok {
		return errors.New("user not found")
	}
	user.Events[event.EventId] = event
	server.UserCalenders[userId] = user
	return nil
}

func (server *CalenderServer) AddUser(userId int) {
	server.Lock()
	defer server.Unlock()
	if _, ok := server.UserCalenders[userId]; !ok {
		user := NewUserCalender(userId)
		server.UserCalenders[userId] = *user
	}
}

func (server *CalenderServer) UpdateEvent(userId int, event Event) error {
	server.Lock()
	defer server.Unlock()
	user, ok := server.UserCalenders[userId]
	if !ok {
		return errors.New("user not found")
	}
	if _, ok = user.Events[event.EventId]; !ok {
		return errors.New("event not found")
	}

	user.Events[event.EventId] = event
	server.UserCalenders[userId] = user
	return nil
}

func (server *CalenderServer) DeleteEvent(userId int, eventId guid.Guid) error {
	server.Lock()
	defer server.Unlock()
	user, ok := server.UserCalenders[userId]
	if !ok {
		return errors.New("user not found")
	}
	delete(user.Events, eventId)
	server.UserCalenders[userId] = user
	return nil
}

func (server *CalenderServer) GetEventsForDay(userId int, day time.Time) ([]Event, error) {
	server.RLock()
	defer server.RUnlock()
	events := make([]Event, 0)
	user, ok := server.UserCalenders[userId]
	if !ok {
		return events, errors.New("user not found")
	}
	for _, v := range user.Events {
		vDay := v.Date
		if vDay == day {
			events = append(events, v)
		}
	}
	return events, nil
}

func (server *CalenderServer) GetEventsForWeek(userId int, day time.Time) ([]Event, error) {
	server.RLock()
	defer server.RUnlock()
	events := make([]Event, 0)
	user, ok := server.UserCalenders[userId]
	if !ok {
		return nil, errors.New("user not found")
	}
	afterWeek := day.AddDate(0, 0, 7)
	for _, v := range user.Events {
		if (v.Date.After(day) || v.Date.Equal(day)) && (v.Date.Before(afterWeek) || v.Date.Equal(afterWeek)) {
			events = append(events, v)
		}
	}
	return events, nil
}

func (server *CalenderServer) GetEventsForMonth(userId int, day time.Time) ([]Event, error) {
	server.RLock()
	defer server.RUnlock()
	events := make([]Event, 0)
	user, ok := server.UserCalenders[userId]
	if !ok {
		return nil, errors.New("user not found")
	}
	afterMonth := day.AddDate(0, 1, 0)
	for _, v := range user.Events {
		if (v.Date.After(day) || v.Date.Equal(day)) && (v.Date.Before(afterMonth) || v.Date.Equal(afterMonth)) {
			events = append(events, v)
		}
	}
	return events, nil
}
