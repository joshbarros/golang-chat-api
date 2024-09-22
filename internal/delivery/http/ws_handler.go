package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joshbarros/golang-chat-api/internal/domain"
	"github.com/joshbarros/golang-chat-api/internal/usecase"
	redis_interface "github.com/joshbarros/golang-chat-api/pkg/db/interfaces"
	"github.com/joshbarros/golang-chat-api/pkg/security"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("websocket-tracer")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	chatUsecase usecase.ChatUsecaseInterface
	redisClient redis_interface.RedisClientInterface
}

func NewWSHandler(
  chatUsecase usecase.ChatUsecaseInterface,
  redisClient redis_interface.RedisClientInterface,
) *WSHandler {
	return &WSHandler{
		chatUsecase: chatUsecase,
		redisClient: redisClient,
	}
}

type CreateRoomRequest struct {
	RoomName string `json:"room_name"`
}

// WebSocketHandler godoc
// @Summary Establish a WebSocket connection
// @Description Connect to a WebSocket for real-time communication in a room
// @Tags websocket
// @Param roomID path string true "Room ID"
// @Produce json
// @Success 101 {string} string "WebSocket Connection Established"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /ws/{roomID} [get]
func (h *WSHandler) WebSocketHandler(c *gin.Context) {
	// Start tracing
	_, span := tracer.Start(c.Request.Context(), "WebSocketHandler")
	defer span.End()

	// Get the JWT token from the Authorization header
	token := c.GetHeader("Authorization")
	if token == "" {
		log.Println("Authorization header is missing")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Strip "Bearer " from the token string
	token = strings.TrimPrefix(token, "Bearer ")

	// Validate the token and extract claims
	claims, err := security.ValidateJWT(token)
	if err != nil {
		log.Println("Invalid token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract userID from the token claims
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID == 0 {
		log.Println("Invalid user ID in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer ws.Close()

	roomID := c.Param("roomID") // Room ID from URL params

	// Check if the room exists in the database
	room, err := h.chatUsecase.GetRoomByID(roomID)
	if err != nil || room == nil {
		log.Printf("Room %s does not exist", roomID)
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Room does not exist"))
		return
	}

	// Room exists, continue to handle messages
	log.Printf("User %d connected to room %s", userID, roomID)

	// Add WebSocket connection to the room
	h.chatUsecase.AddClientToRoom(roomID, ws)

	done := make(chan bool)

	// Start broadcasting messages for the room
	go h.chatUsecase.BroadcastMessages(roomID, done)

	// Handle incoming messages
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Create a message object, with the userID extracted from the token
		msg := domain.Message{
			UserID:    userID,
			RoomID:    roomID,
			Message:   string(message),
			Timestamp: time.Now(),
		}

		// Send the message to the worker pool
		if err := h.chatUsecase.SendMessageToRoom(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}

	// Remove WebSocket connection from the room
	h.chatUsecase.RemoveClientFromRoom(roomID, ws)

	// Close the room when the connection is closed
	h.chatUsecase.CloseRoom(roomID, done)
}

// GetRooms godoc
// @Summary Get a list of available chat rooms
// @Description Retrieve all available rooms for users to join
// @Tags rooms
// @Produce  json
// @Success 200 {array} domain.Room
// @Failure 500 {object} map[string]string
// @Router /rooms [get]
func (h *WSHandler) GetRooms(c *gin.Context) {
  rooms, err := h.chatUsecase.GetAvailableRooms()
  if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch rooms"})
      return
  }
  c.JSON(http.StatusOK, rooms)
}

// CreateRoom godoc
// @Summary Create a new chat room
// @Description Create a room for users to join
// @Tags rooms
// @Accept json
// @Produce json
// @Param room body CreateRoomRequest true "Room Info"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rooms [post]
func (h *WSHandler) CreateRoom(c *gin.Context) {
  var req struct {
      RoomName string `json:"room_name"`
  }

  if err := c.ShouldBindJSON(&req); err != nil || req.RoomName == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room data"})
      return
  }

  // Create the room object (without the ID, which will be auto-generated)
  room := &domain.Room{
      RoomName: req.RoomName,
  }

  // Create the room in the database
  if err := h.chatUsecase.CreateRoom(room); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create room"})
      return
  }

  c.JSON(http.StatusOK, gin.H{"message": "Room created", "room_id": room.ID})
}

// GetRoomMessages godoc
// @Summary Get messages from a specific chat room
// @Description Fetch the last 50 messages from a specified room
// @Tags messages
// @Produce  json
// @Param roomID path string true "Room ID"
// @Success 200 {array} domain.Message
// @Failure 500 {object} map[string]string
// @Router /rooms/{roomID}/messages [get]
func (h *WSHandler) GetRoomMessages(c *gin.Context) {
  roomID := c.Param("roomID")

  // Fetch last 50 messages for the room
  messages, err := h.chatUsecase.GetMessagesByRoom(roomID, 50)
  if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch messages"})
      return
  }

  c.JSON(http.StatusOK, messages)
}
