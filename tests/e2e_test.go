package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/joshua468/user-authentication/controllers"
	"github.com/joshua468/user-authentication/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.Organisation{})

	r := gin.Default()

	authController := controllers.NewAuthController(db)
	orgController := controllers.NewOrganisationController(db)
	userController := controllers.NewUserController(db)

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authController.Register)
			authRoutes.POST("/login", authController.Login)
		}
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("/:id", userController.GetUser)
		}
		orgRoutes := api.Group("/organisations")
		{
			orgRoutes.GET("/", orgController.GetOrganisations)
			orgRoutes.GET("/:orgId", orgController.GetOrganisation)
			orgRoutes.POST("/", orgController.CreateOrganisation)
			orgRoutes.POST("/:orgId/users", orgController.AddUserToOrganisation)
		}
	}

	return r
}

func TestRegister(t *testing.T) {
	router := setupRouter()

	// Test successful registration
	user := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Phone:     "1234567890",
	}

	jsonUser, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Test missing fields
	invalidUser := models.User{
		Email:    "invalid@example.com",
		Password: "password123",
	}

	jsonInvalidUser, _ := json.Marshal(invalidUser)

	req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonInvalidUser))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
