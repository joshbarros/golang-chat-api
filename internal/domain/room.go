package domain

import "time"

type Room struct {
  ID        string    `json:"id"`
  RoomName  string    `json:"room_name"`
  CreatedAt time.Time `json:"created_at"`
}
