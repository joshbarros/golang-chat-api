package usecase

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joshbarros/golang-chat-api/internal/domain"
	"github.com/joshbarros/golang-chat-api/internal/repository"
	"github.com/joshbarros/golang-chat-api/internal/workerpool"
)

type ChatUsecaseInterface interface {
	SendMessageToRoom(msg domain.Message) error
	BroadcastMessages(roomID string, done chan bool)
	CreateRoom(room *domain.Room) error
	CloseRoom(roomID string, done chan bool)
	GetMessagesByRoom(roomID string, limit int) ([]domain.Message, error)
	GetAvailableRooms() ([]domain.Room, error)
  GetRoomByID(roomID string) (*domain.Room, error)
  AddClientToRoom(roomID string, client *websocket.Conn)
  RemoveClientFromRoom(roomID string, client *websocket.Conn)
  GetConnectedClients(roomID string) []*websocket.Conn
}

type ChatUsecase struct {
	messageRepo *repository.MessageRepository
	roomRepo    *repository.RoomRepository
	rooms       map[string]chan domain.Message
  clients     map[string][]*websocket.Conn
	roomsMutex  sync.RWMutex
	workerPool  *workerpool.WorkerPool
}

func NewChatUsecase(
	messageRepo *repository.MessageRepository,
	roomRepo *repository.RoomRepository,
	workerPool *workerpool.WorkerPool,
) *ChatUsecase {
	return &ChatUsecase{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		rooms:       make(map[string]chan domain.Message),
    clients:     make(map[string][]*websocket.Conn),
		workerPool:  workerPool,
	}
}

// Add the actual implementation of GetRoomByID
func (uc *ChatUsecase) GetRoomByID(roomID string) (*domain.Room, error) {
  // Call the repository function to get the room from the database
  room, err := uc.roomRepo.GetRoomByID(roomID)
  if err != nil {
      return nil, fmt.Errorf("error fetching room by ID: %w", err)
  }
  return room, nil
}

func (uc *ChatUsecase) GetAvailableRooms() ([]domain.Room, error) {
  // Fetch rooms from the PostgreSQL repository instead of the in-memory map
  rooms, err := uc.roomRepo.GetRooms()
  if err != nil {
      return nil, fmt.Errorf("error fetching rooms from database: %w", err)
  }

  return rooms, nil
}

func (uc *ChatUsecase) GetMessagesByRoom(roomID string, limit int) ([]domain.Message, error) {
  return uc.messageRepo.GetMessagesByRoom(roomID, limit)
}

// Send message to a room
func (uc *ChatUsecase) SendMessageToRoom(msg domain.Message) error {
	uc.workerPool.AddJob(msg)
	log.Printf("Message sent to worker pool for room: %s", msg.RoomID)
	return nil
}

func (uc *ChatUsecase) BroadcastMessages(roomID string, done chan bool) {
  uc.roomsMutex.RLock()
  room, exists := uc.rooms[roomID]
  uc.roomsMutex.RUnlock()

  if !exists {
      log.Printf("Room %s does not exist", roomID)
      return
  }

  // Continuously listen for messages in the room and broadcast them
  for {
      select {
      case msg := <-room:
          // Save message to DB
          if err := uc.messageRepo.SaveMessage(msg); err != nil {
              log.Printf("Error saving message: %v", err)
          }

          // Broadcast message to all clients in the room
          for _, client := range uc.GetConnectedClients(roomID) {
              if client != nil {
                  if err := client.WriteJSON(msg); err != nil {
                      log.Printf("Error broadcasting message to client: %v", err)
                  }
              }
          }

          log.Printf("Message broadcasted: %s", msg.Message)
      case <-done:
          log.Printf("Shutting down room %s", roomID)
          return
      }
  }
}

func (uc *ChatUsecase) CreateRoom(room *domain.Room) error {
  uc.roomsMutex.Lock()
  defer uc.roomsMutex.Unlock()

  // Check if the room already exists in the database by room name
  dbRoom, err := uc.roomRepo.GetRoomByName(room.RoomName)
  if err != nil {
      return fmt.Errorf("error checking room in database: %w", err)
  }

  if dbRoom != nil {
      log.Printf("Room %s already exists in the database", room.RoomName)
      return fmt.Errorf("room already exists")
  }

  // Create a new room in the database
  err = uc.roomRepo.CreateRoom(room) // This should generate the ID
  if err != nil {
      return fmt.Errorf("error creating room in the database: %w", err)
  }

  // Now that the room has an ID, create the room in memory
  uc.rooms[room.ID] = make(chan domain.Message)
  log.Printf("Room %s created with ID %s", room.RoomName, room.ID)
  return nil
}

// Close the room
func (uc *ChatUsecase) CloseRoom(roomID string, done chan bool) {
	uc.roomsMutex.Lock()
	defer uc.roomsMutex.Unlock()

	if room, exists := uc.rooms[roomID]; exists {
		close(room)  // Close the message channel
		delete(uc.rooms, roomID)  // Remove the room from the map
		log.Printf("Room %s closed", roomID)
		done <- true
	} else {
		log.Printf("Room %s does not exist", roomID)
		done <- false
	}
}

// getConnectedClients returns all clients connected to a specific room
func (uc *ChatUsecase) GetConnectedClients(roomID string) []*websocket.Conn {
  // Return a list of connected WebSocket clients in the room
  uc.roomsMutex.RLock()
  defer uc.roomsMutex.RUnlock()

  if clients, exists := uc.clients[roomID]; exists {
      return clients
  }

  return nil
}

// AddClientToRoom adds a WebSocket connection to a room
func (uc *ChatUsecase) AddClientToRoom(roomID string, client *websocket.Conn) {
  uc.roomsMutex.Lock()
  defer uc.roomsMutex.Unlock()

  // Initialize room if not already present
  if _, exists := uc.clients[roomID]; !exists {
      uc.clients[roomID] = []*websocket.Conn{}
  }

  // Add the WebSocket client to the room
  uc.clients[roomID] = append(uc.clients[roomID], client)
  log.Printf("Client added to room %s", roomID)
}

// RemoveClientFromRoom removes a WebSocket connection from a room
func (uc *ChatUsecase) RemoveClientFromRoom(roomID string, client *websocket.Conn) {
  uc.roomsMutex.Lock()
  defer uc.roomsMutex.Unlock()

  if clients, exists := uc.clients[roomID]; exists {
      for i, c := range clients {
          if c == client {
              // Remove the client from the slice
              uc.clients[roomID] = append(clients[:i], clients[i+1:]...)
              log.Printf("Client removed from room %s", roomID)
              break
          }
      }
  }
}
