package notify

import "time"

type Notification struct {
	Message  string    `json:"message"`
	DateTime time.Time `json:"datetime"`
}
