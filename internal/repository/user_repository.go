package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/joshbarros/golang-chat-api/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	// Debugging log (avoid logging sensitive information in production)
	log.Printf("Inserting user with username: %s", user.Username)

	// Execute the query and scan the generated user ID
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("error creating user %s: %w", user.Username, err)
	}

	log.Printf("User created with ID: %d", user.ID)
	return nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, username, email, password
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("error retrieving user by email %s: %w", email, err)
	}

	return &user, nil
}
