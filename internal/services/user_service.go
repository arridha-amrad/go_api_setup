package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"my-go-api/internal/dto"
	"my-go-api/internal/models"
	"my-go-api/internal/repositories"
	"my-go-api/pkg/utils"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	user, err := u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (u *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user, err := u.userRepo.GetById(ctx, userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return u.userRepo.GetAll(ctx)
}

func (u *UserService) CreateUser(ctx context.Context, req dto.CreateUser) (*models.User, error) {
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username is registered")
	}

	existingUser, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email is registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.Create(ctx, req.Name, req.Username, req.Email, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("create user failed")
	}

	return user, nil
}
