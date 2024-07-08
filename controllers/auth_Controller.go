package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/joshua468/user-authentication/models"
	"github.com/joshua468/user-authentication/utils"
)

type AuthController struct {
	DB        *gorm.DB
	JWTSecret string
}

func NewAuthController(db *gorm.DB, jwtSecret string) *AuthController {
	return &AuthController{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var input struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		UserID:    utils.GenerateUUID(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hashedPassword),
		Phone:     input.Phone,
	}

	if err := ctrl.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.UserID, ctrl.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registration successful",
		"data": gin.H{
			"accessToken": token,
			"user": gin.H{
				"userId":    user.UserID,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"email":     user.Email,
				"phone":     user.Phone,
			},
		},
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": err.Error()})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.UserID, ctrl.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"accessToken": token,
			"user": gin.H{
				"userId":    user.UserID,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"email":     user.Email,
				"phone":     user.Phone,
			},
		},
	})
}
