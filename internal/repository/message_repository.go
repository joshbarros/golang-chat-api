package repository

import (
	"database/sql"
	"fmt"

	"github.com/joshbarros/golang-chat-api/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveMessage inserts a message into the database
func (r *MessageRepository) SaveMessage(msg domain.Message) error {
  // Check if the room exists before saving the message
  roomQuery := `SELECT COUNT(1) FROM rooms WHERE id = $1`
  var roomCount int
  err := r.db.QueryRow(roomQuery, msg.RoomID).Scan(&roomCount)
  if err != nil || roomCount == 0 {
      return fmt.Errorf("room %s does not exist", msg.RoomID)
  }

	query := `
		INSERT INTO messages (user_id, room_id, message, timestamp)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err = r.db.QueryRow(query, msg.UserID, msg.RoomID, msg.Message, msg.Timestamp).Scan(&msg.ID)
	if err != nil {
		return fmt.Errorf("error saving message for room %s: %w", msg.RoomID, err)
	}
	return nil
}

// GetMessagesByRoom fetches the last 'limit' messages for a given room
func (r *MessageRepository) GetMessagesByRoom(roomID string, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	query := `
		SELECT id, user_id, room_id, message, timestamp
		FROM messages
		WHERE room_id=$1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	rows, err := r.db.Query(query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching messages for room %s: %w", roomID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.RoomID, &msg.Message, &msg.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning message for room %s: %w", roomID, err)
		}
		messages = append(messages, msg)
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return messages, nil
}
