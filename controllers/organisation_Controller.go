package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/joshua468/user-authentication/models"
)

type OrganisationController struct {
	db *gorm.DB
}

func NewOrganisationController(db *gorm.DB) *OrganisationController {
	return &OrganisationController{db}
}

func (oc *OrganisationController) GetOrganisations(c *gin.Context) {
	var orgs []models.Organisation
	userId := c.MustGet("userId").(string)

	if err := oc.db.Joins("JOIN organisation_users on organisation_users.organisation_id = organisations.id").
		Joins("JOIN users on users.id = organisation_users.user_id").
		Where("users.user_id = ?", userId).
		Find(&orgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve organisations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Organisations retrieved",
		"data":    orgs,
	})
}

func (oc *OrganisationController) GetOrganisation(c *gin.Context) {
	orgId := c.Param("orgId")
	var org models.Organisation

	if err := oc.db.Where("org_id = ?", orgId).First(&org).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organisation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Organisation found",
		"data":    org,
	})
}

func (oc *OrganisationController) CreateOrganisation(c *gin.Context) {
	var org models.Organisation

	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org.OrgID = uuid.New().String()

	if err := oc.db.Create(&org).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organisation"})
		return
	}

	userId := c.MustGet("userId").(string)
	var user models.User
	if err := oc.db.Where("user_id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	if err := oc.db.Model(&org).Association("Users").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate user with organisation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Organisation created successfully",
		"data":    org,
	})
}

func (oc *OrganisationController) AddUserToOrganisation(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgId := c.Param("orgId")
	var org models.Organisation

	if err := oc.db.Where("org_id = ?", orgId).First(&org).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organisation not found"})
		return
	}

	var user models.User
	if err := oc.db.Where("user_id = ?", input.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := oc.db.Model(&org).Association("Users").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to organisation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User added to organisation successfully",
	})
}
