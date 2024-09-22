package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_http "github.com/joshbarros/golang-chat-api/internal/delivery/http"
	"github.com/joshbarros/golang-chat-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase for testing
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) Register(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserUsecase) Login(email, password string) (*domain.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func mockJWTGenerator(userID string) (string, error) {
	return "mocked-token", nil
}

// Test case for Register
func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock the usecase and setup the handler
	mockUsecase := new(MockUserUsecase)
	userHandler := _http.NewUserHandler(mockUsecase)

	router := gin.Default()
	router.POST("/register", userHandler.Register)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful registration",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "testuser@example.com",
				"password": "password123",
			},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"User registered successfully"}`,
		},
		{
			name: "Missing fields",
			requestBody: map[string]string{
				"username": "",
				"email":    "",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Username, email, and password are required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase.On("Register", mock.Anything).Return(tt.mockReturnErr)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockUsecase.AssertExpectations(t)
		})
	}
}

// Test case for Login
func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock the usecase and setup the handler with mock JWT generator
	mockUsecase := new(MockUserUsecase)
	userHandler := _http.NewUserHandler(mockUsecase)

	router := gin.Default()
	router.POST("/login", userHandler.Login)

	tests := []struct {
		name           string
		requestBody    map[string]string
		mockReturnUser *domain.User
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful login",
			requestBody: map[string]string{
				"email":    "testuser@example.com",
				"password": "password123",
			},
			mockReturnUser: &domain.User{ID: 1, Email: "testuser@example.com"},
			mockReturnErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"mocked-token"}`,
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]string{
				"email":    "wronguser@example.com",
				"password": "wrongpassword",
			},
			mockReturnUser: nil,
			mockReturnErr:  errors.New("Invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid credentials"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase.On("Login", tt.requestBody["email"], tt.requestBody["password"]).Return(tt.mockReturnUser, tt.mockReturnErr)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockUsecase.AssertExpectations(t)
		})
	}
}
