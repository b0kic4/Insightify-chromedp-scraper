package models

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username      string `gorm:"index:,unique,where:username <> ''"` // Unique if not empty
	FullName      string
	Email         string `gorm:"index:,unique,where:email <> ''"` // Unique if not empty
	PasswordHash  string
	Provider      string
	ProviderID    string
	AvatarURL     string
	VerifiedEmail bool
}

type Server struct {
	db *gorm.DB
}

func CreateUser(db *gorm.DB, user User) error {
	// Check if the user exists depending on the authentication method
	var existingUser User
	var err error
	if user.Provider != "" && user.ProviderID != "" {
		// OAuth registration
		err = db.Where("provider = ? AND provider_id = ?", user.Provider, user.ProviderID).First(&existingUser).Error
	} else if user.Email != "" {
		// Email based registration
		err = db.Where("email = ?", user.Email).First(&existingUser).Error
	} else if user.Username != "" {
		// Username based registration (if applicable)
		err = db.Where("username = ?", user.Username).First(&existingUser).Error
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existingUser.ID != 0 {
		return fmt.Errorf("user already exists")
	}

	// No existing user found, proceed to create a new one
	result := db.Create(&user)
	return result.Error
}
