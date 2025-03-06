package dto

type CreateUser struct {
	Name     string `json:"name" validate:"required,min=5"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=5"`
	Password string `json:"password" validate:"required,strongPassword"`
}
