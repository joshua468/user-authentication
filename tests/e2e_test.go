package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/joshua468/user-authentication/controllers"
	"github.com/joshua468/user-authentication/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter() (*gin.Engine, *gorm.DB) {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Error connecting to database")
	}
	db.AutoMigrate(&models.User{}, &models.Organisation{}) // Adjust migrations as needed

	r := gin.Default()

	// Use environment variable for JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")

	authController := controllers.NewAuthController(db, jwtSecret)
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
			userRoutes.GET("/", func(c *gin.Context) {
				var users []models.User
				db.Find(&users)
				c.JSON(http.StatusOK, gin.H{"users": users})
			})
		}
		orgRoutes := api.Group("/organisations")
		{
			orgRoutes.GET("/", orgController.GetOrganisations)
			orgRoutes.GET("/:orgId", orgController.GetOrganisation)
			orgRoutes.POST("/", orgController.CreateOrganisation)
			orgRoutes.POST("/:orgId/users", orgController.AddUserToOrganisation)
		}
	}

	return r, db
}

func registerUser(router *gin.Engine, user models.User) map[string]interface{} {
	jsonUser, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	return response
}

func TestRegisterUser(t *testing.T) {
	router, _ := setupRouter()

	user := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Phone:     "1234567890",
	}

	response := registerUser(router, user)

	assert.Equal(t, http.StatusCreated, int(response["code"].(float64)))

	// Check registration response
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Registration successful", response["message"])

	data := response["data"].(map[string]interface{})
	accessToken := data["accessToken"].(string)
	registeredUser := data["user"].(map[string]interface{})

	assert.Equal(t, user.FirstName, registeredUser["firstName"])
	assert.Equal(t, user.LastName, registeredUser["lastName"])
	assert.Equal(t, user.Email, registeredUser["email"])
	assert.NotEmpty(t, accessToken)
}

func TestLoginUserSuccessfully(t *testing.T) {
	router, _ := setupRouter()

	user := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Phone:     "1234567890",
	}

	// Register the user first
	registerResponse := registerUser(router, user)
	assert.Equal(t, http.StatusCreated, int(registerResponse["code"].(float64)))

	// Check registration response
	assert.Equal(t, "success", registerResponse["status"])
	assert.Equal(t, "Registration successful", registerResponse["message"])

	data := registerResponse["data"].(map[string]interface{})
	accessToken := data["accessToken"].(string)
	registeredUser := data["user"].(map[string]interface{})

	assert.Equal(t, user.FirstName, registeredUser["firstName"])
	assert.Equal(t, user.LastName, registeredUser["lastName"])
	assert.Equal(t, user.Email, registeredUser["email"])
	assert.NotEmpty(t, accessToken)

	// Now, attempt to log in with valid credentials
	loginDetails := map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}

	jsonLoginDetails, _ := json.Marshal(loginDetails)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonLoginDetails))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check login response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Login successful", response["message"])

	loginData := response["data"].(map[string]interface{})
	loginAccessToken := loginData["accessToken"].(string)
	loggedInUser := loginData["user"].(map[string]interface{})

	assert.Equal(t, user.FirstName, loggedInUser["firstName"])
	assert.Equal(t, user.LastName, loggedInUser["lastName"])
	assert.Equal(t, user.Email, loggedInUser["email"])
	assert.NotEmpty(t, loginAccessToken)
}

func TestGetAllUsers(t *testing.T) {
	router, db := setupRouter()

	// Register a few users
	users := []models.User{
		{
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice.smith@example.com",
			Password:  "password123",
			Phone:     "1234567890",
		},
		{
			FirstName: "Bob",
			LastName:  "Johnson",
			Email:     "bob.johnson@example.com",
			Password:  "password456",
			Phone:     "0987654321",
		},
	}

	for _, user := range users {
		registerUser(router, user)
	}

	// Retrieve all users
	req, _ := http.NewRequest("GET", "/api/users/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(response)) // Assuming two users were registered

	// Optionally, you can add more detailed assertions on the returned user data
	for _, user := range response {
		assert.Contains(t, user["email"], "@example.com")
	}

	// Clean up: Close the database connection
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Error getting underlying DB connection: %v", err)
	}
	sqlDB.Close()
}

func TestRegisterUserMissingRequiredFields(t *testing.T) {
	router, _ := setupRouter()

	// Test missing firstName
	invalidUser := models.User{
		LastName: "Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
		Phone:    "1234567890",
	}

	jsonInvalidUser, _ := json.Marshal(invalidUser)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonInvalidUser))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	// Check response body for error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Contains(t, response["errors"], "firstName")

	// Test missing email
	invalidUser = models.User{
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
		Phone:     "1234567890",
	}

	jsonInvalidUser, _ = json.Marshal(invalidUser)

	req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonInvalidUser))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	// Check response body for error message
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Contains(t, response["errors"], "email")

	// Repeat similar tests for other required fields
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	router, _ := setupRouter()

	// Register a user with a valid email first
	validUser := models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		Phone:     "1234567890",
	}

	registerUser(router, validUser)

	// Attempt to register another user with the same email
	duplicateUser := models.User{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "john.doe@example.com", // Same as the previously registered user
		Password:  "password456",
		Phone:     "0987654321",
	}

	jsonDuplicateUser, _ := json.Marshal(duplicateUser)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonDuplicateUser))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check response body for error message
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)

	assert.Contains(t, response["error"], "Failed to register user")
	
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	code := m.Run()
	os.Exit(code)
}
