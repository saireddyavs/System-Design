package services

import (
	"context"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// PreferenceService manages user notification preferences
type PreferenceService struct {
	userRepo interfaces.UserRepository
}

// NewPreferenceService creates a new preference service
func NewPreferenceService(userRepo interfaces.UserRepository) *PreferenceService {
	return &PreferenceService{
		userRepo: userRepo,
	}
}

// GetUserPreferences returns the user's channel preferences
func (s *PreferenceService) GetUserPreferences(ctx context.Context, userID string) (map[models.Channel]bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	prefs := make(map[models.Channel]bool)
	for ch, enabled := range user.Preferences {
		prefs[ch] = enabled
	}
	return prefs, nil
}

// SetChannelPreference updates a user's preference for a channel
func (s *PreferenceService) SetChannelPreference(ctx context.Context, userID string, channel models.Channel, enabled bool) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	user.SetChannelPreference(channel, enabled)
	return s.userRepo.Save(ctx, user)
}

// GetEnabledChannels returns channels the user has enabled
func (s *PreferenceService) GetEnabledChannels(ctx context.Context, userID string) ([]models.Channel, error) {
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	var enabled []models.Channel
	for ch, ok := range prefs {
		if ok {
			enabled = append(enabled, ch)
		}
	}
	return enabled, nil
}
