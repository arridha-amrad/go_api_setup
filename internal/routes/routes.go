package routes

import (
	"database/sql"
	"my-go-api/internal/handlers"
	"my-go-api/internal/middleware"

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
	userHandler := handlers.NewUserHandler(userService)

	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, userService)

	md := middleware.RegisterValidationMiddleware(validate)
	mdT := middleware.RegisterTokenVerificationMiddleware(authService)

	router.SetTrustedProxies([]string{"127.0.0.1"})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to V1"})
		})
		v1Users := v1.Group("/users")
		{
			v1Users.GET("", userHandler.GetAll)
			v1Users.POST("", md.CreateUser, userHandler.Create)
			v1Users.GET("/:id", userHandler.GetUserById)
			v1Users.PUT("/:id", md.UpdateUser, userHandler.Update)
		}
		v1Auth := v1.Group("/auth")
		{
			v1Auth.GET("", mdT.RequireAuth, authHandler.GetAuth)
			v1Auth.POST("", md.Login, authHandler.Login)
		}
	}

	return router
}
