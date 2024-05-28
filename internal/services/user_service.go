package services

import (
	"Insightify-backend/internal/database/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUserOrUpdate(ctx context.Context, user models.User) (*models.User, error) {
	// Check if the user already exists
	var existingUser models.User
	err := s.db.WithContext(ctx).Where("provider_id = ? AND provider = ?", user.ProviderID, user.Provider).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// If not found, create a new user
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// If found, update the existing user
		user.ID = existingUser.ID // Preserve the existing user ID
		if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
	}
	return &user, nil
}
