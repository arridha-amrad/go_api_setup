package routes

import (
	"database/sql"
	"my-go-api/internal/dto"
	"my-go-api/internal/handlers"

	"my-go-api/internal/repositories"
	"my-go-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RegisterRoutes(db *sql.DB, validate *validator.Validate) *gin.Engine {
	router := gin.Default()

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService, validate)

	userValidation := dto.RegisterUserValidationMiddleware(validate)

	router.SetTrustedProxies([]string{"127.0.0.1"})
	v1 := router.Group("/api/v1")
	{
		v1.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"message": "Welcome"})
		})
		v1Users := v1.Group("/users")
		{
			v1Users.GET("", userHandler.GetAll)
			v1Users.POST("", userValidation.CreateUser, userHandler.Create)
		}
	}

	return router
}
