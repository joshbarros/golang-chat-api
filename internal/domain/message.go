package domain

import "time"

type Message struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    RoomID    string    `json:"room_id"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
