package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/joshua468/user-authentication/controllers"
	"github.com/joshua468/user-authentication/middlewares"
	"github.com/joshua468/user-authentication/models"
)

var db *gorm.DB

func connect() {

	dbhost := os.Getenv("DB_HOST")
	dbuser := os.Getenv("DB_USER")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("PORT")

	// Database connection
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", dbhost, dbuser, dbpassword, dbname, port)
	fmt.Println(dsn)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{}, &models.Organisation{})
}

// func loadenv() {
// 	if err := godotenv.Load(".env"); err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}
// }

func loadserver() {
	router := gin.Default()

	// Initialize controllers
	authController := controllers.NewAuthController(db, os.Getenv("JWT_SECRET"))
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
		userRoutes := api.Group("/users").Use(middlewares.JWTAuthMiddleware(os.Getenv("JWT_SECRET")))
		{
			userRoutes.GET("/:id", userController.GetUser)
		}
		orgRoutes := api.Group("/organisations").Use(middlewares.JWTAuthMiddleware(os.Getenv("JWT_SECRET")))
		{
			orgRoutes.GET("/", orgController.GetOrganisations)
			orgRoutes.GET("/:orgId", orgController.GetOrganisation)
			orgRoutes.POST("/", orgController.CreateOrganisation)
			orgRoutes.POST("/:orgId/users", orgController.AddUserToOrganisation)
		}
	}

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func main() {
	// loadenv()
	connect()
	loadserver()

}
