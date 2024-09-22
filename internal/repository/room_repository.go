package repository

import (
	"database/sql"
	"fmt"

	"github.com/joshbarros/golang-chat-api/internal/domain"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// CreateRoom inserts a new room into the database and returns the generated ID.
func (r *RoomRepository) CreateRoom(room *domain.Room) error {
	query := `
		INSERT INTO rooms (room_name)
		VALUES ($1)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, room.RoomName).Scan(&room.ID, &room.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating room %s: %w", room.RoomName, err)
	}

	return nil
}

// GetRoomByID retrieves a room by its ID
func (r *RoomRepository) GetRoomByID(roomID string) (*domain.Room, error) {
	var room domain.Room
	query := `
		SELECT id, room_name, created_at
		FROM rooms
		WHERE id = $1
	`

	err := r.db.QueryRow(query, roomID).Scan(&room.ID, &room.RoomName, &room.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found with id: %s", roomID)
		}
		return nil, fmt.Errorf("error retrieving room by id %s: %w", roomID, err)
	}

	return &room, nil
}

// GetRoomByName retrieves a room by its name
func (r *RoomRepository) GetRoomByName(roomName string) (*domain.Room, error) {
	var room domain.Room
	query := `
		SELECT id, room_name, created_at
		FROM rooms
		WHERE room_name = $1
	`

	err := r.db.QueryRow(query, roomName).Scan(&room.ID, &room.RoomName, &room.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no room is found
		}
		return nil, fmt.Errorf("error retrieving room by name %s: %w", roomName, err)
	}

	return &room, nil
}


// GetRooms retrieves all available rooms
func (r *RoomRepository) GetRooms() ([]domain.Room, error) {
	var rooms []domain.Room
	query := `
		SELECT id, room_name, created_at
		FROM rooms
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching rooms: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.RoomName, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return rooms, nil
}
