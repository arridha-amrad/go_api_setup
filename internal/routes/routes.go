package routes

import (
	"database/sql"
	"my-go-api/internal/config"
	"my-go-api/internal/handlers"
	"my-go-api/internal/middleware"
	"my-go-api/internal/utils"

	"my-go-api/internal/repositories"
	"my-go-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RegisterRoutes(db *sql.DB, validate *validator.Validate, config *config.Config) *gin.Engine {
	router := gin.Default()

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	utilities := utils.NewUtilities(config.JWtSecretKey, config.AppUri, config.GoogleOAuth2)
	tokenRepo := repositories.NewTokenRepository(db)

	authService := services.NewAuthService(
		userRepo,
		utilities,
		tokenRepo,
		config.AppUri,
	)

	authHandler := handlers.NewAuthHandler(authService, userService)

	md := middleware.RegisterValidationMiddleware(validate)
	mdT := middleware.RegisterTokenVerificationMiddleware(authService)

	router.SetTrustedProxies([]string{"127.0.0.1"})

	v1 := router.Group("/api/v1")
	{
		v1.GET("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to V1"})
		})
		v1Users := v1.Group("/users")
		{
			v1Users.GET("", userHandler.GetAll)
			v1Users.GET("/:id", userHandler.GetUserById)
			v1Users.PUT("/:id", md.UpdateUser, userHandler.Update)
		}
		v1Auth := v1.Group("/auth")
		{
			v1Auth.GET("", mdT.RequireAuth, authHandler.GetAuth)
			v1Auth.POST("", md.Login, authHandler.Login)
			v1Auth.POST("/refresh-token", authHandler.RefreshToken)
			v1Auth.POST("/logout", authHandler.Logout)
			v1Auth.POST("/register", md.CreateUser, authHandler.Register)
		}
	}

	return router
}
