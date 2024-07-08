package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/joshua468/user-authentication/models"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db}
}

func (uc *UserController) GetUser(c *gin.Context) {
	userId := c.Param("id")
	var user models.User
	if err := uc.db.Where("user_id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User found",
		"data": gin.H{
			"userId":    user.UserID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"email":     user.Email,
			"phone":     user.Phone,
		},
	})
}
