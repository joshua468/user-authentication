package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joshua468/user-authentication/config"
	"github.com/joshua468/user-authentication/controllers"
	"github.com/joshua468/user-authentication/middlewares"
	"github.com/joshua468/user-authentication/models"
)

var (
	db  *gorm.DB
	cfg config.Config
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load configuration
	cfg = config.LoadConfig()

	// Database connection
	dsn := "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Organisation{})
}

// Handler is the exported function for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set up Gin
	router := gin.Default()

	// Initialize controllers
	authController := controllers.NewAuthController(db, cfg.JWTSecret)
	orgController := controllers.NewOrganisationController(db)
	userController := controllers.NewUserController(db)

	// Routes
	api := router.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authController.Register)
			authRoutes.POST("/login", authController.Login)
		}
		userRoutes := api.Group("/users").Use(middlewares.JWTAuthMiddleware(cfg.JWTSecret))
		{
			userRoutes.GET("/:id", userController.GetUser)
		}
		orgRoutes := api.Group("/organisations").Use(middlewares.JWTAuthMiddleware(cfg.JWTSecret))
		{
			orgRoutes.GET("/", orgController.GetOrganisations)
			orgRoutes.GET("/:orgId", orgController.GetOrganisation)
			orgRoutes.POST("/", orgController.CreateOrganisation)
			orgRoutes.POST("/:orgId/users", orgController.AddUserToOrganisation)
		}
	}

	// Serve the request
	router.ServeHTTP(w, r)
}
