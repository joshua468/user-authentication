package models

import "gorm.io/gorm"

type Organisation struct {
	gorm.Model
	OrgID       string `gorm:"unique;not null" json:"orgId"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Users       []User `gorm:"many2many:organisation_users;" json:"users"`
}
