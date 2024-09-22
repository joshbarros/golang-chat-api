package usecase

import (
	"errors"

	"github.com/joshbarros/golang-chat-api/internal/domain"
	"github.com/joshbarros/golang-chat-api/internal/repository"
	"github.com/joshbarros/golang-chat-api/pkg/security"
)

type UserUsecaseInterface interface {
	Register(user *domain.User) error
	Login(email, password string) (*domain.User, error)
}

type UserUsecase struct {
	userRepo *repository.UserRepository
}

func NewUserUsecase(userRepo *repository.UserRepository) *UserUsecase {
	return &UserUsecase{userRepo: userRepo}
}

func (uc *UserUsecase) Register(user *domain.User) error {
	// Validate user input and add business logic if needed
	return uc.userRepo.CreateUser(user)
}

func (uc *UserUsecase) Login(email, password string) (*domain.User, error) {
	// Fetch the user by email
	user, err := uc.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if the password is correct
	if err := security.CheckPasswordHash(password, user.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	// Return the authenticated user
	return user, nil
}
