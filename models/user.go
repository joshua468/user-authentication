package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	UserID    string `gorm:"unique;not null" json:"userId"`
	FirstName string `gorm:"not null" json:"firstName"`
	LastName  string `gorm:"not null" json:"lastName"`
	Email     string `gorm:"unique;not null" json:"email"`
	Password  string `gorm:"not null" json:"password"`
	Phone     string `json:"phone"`
}
