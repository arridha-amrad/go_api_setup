package services

import (
	"context"
	"my-go-api/internal/models"
	"my-go-api/internal/repositories"

	"github.com/google/uuid"
)

type IUserService interface {
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userService struct {
	userRepo repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) IUserService {
	return &userService{userRepo: userRepo}
}

func (u *userService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user, err := u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (u *userService) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user, err := u.userRepo.GetById(ctx, userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return u.userRepo.GetAll(ctx)
}
