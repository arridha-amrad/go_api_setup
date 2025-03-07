package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"my-go-api/internal/models"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (s *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, name, username, email, provider, role, created_at, updated_at
		FROM users
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []models.User{}
	// Iterate over the rows
	for rows.Next() {
		var user models.User

		// Scan the row into the User struct
		err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Provider, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Failed to scan user: %v", err) // Log the error
			return nil, err
		}

		// Append the user to the slice
		users = append(users, user)
	}
	return users, nil

}

func (s *UserRepository) Create(ctx context.Context, name, username, email, password string) (*models.User, error) {
	fmt.Println("name : ", name)
	user := &models.User{}
	query := `
						INSERT INTO users (name, username, email, password)
						VALUES ($1, $2, $3, $4)
						RETURNING id, name, email, username, provider, role, updated_at, created_at
					`
	if err := s.db.QueryRowContext(ctx, query, name, username, email, password).Scan(&user.ID, &user.Name, &user.Email, &user.Username, &user.Provider, &user.Role, &user.UpdatedAt, &user.CreatedAt); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserRepository) GetById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT 
						id, name, username, email, password, provider, role, created_at, updated_at 
						FROM users WHERE id = $1`
	if err := s.db.QueryRowContext(ctx, query, userId).Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.Provider, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT 
						id, name, username, email, password, provider, role, created_at, updated_at 
						FROM users WHERE username = $1`
	if err := s.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.Provider, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT 
						id, name, username, email, password, provider, role, created_at, updated_at 
						FROM users WHERE email = $1`
	if err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.Provider, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		UPDATE users
		SET username=$1, email=$2, name=$3, password=$4, role=$5
		WHERE id=$6 
		RETURNING id, name, username, email, password, provider, role, created_at, updated_at
	`
	if err := s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Name, user.Password, user.Role, user.ID).Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.Provider, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return user, nil
}
