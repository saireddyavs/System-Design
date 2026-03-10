package repositories

import (
	"errors"
	"strings"
	"sync"

	"social-media-platform/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrFriendshipNotFound = errors.New("friendship not found")
)

// InMemoryUserRepository implements UserRepository with thread-safe in-memory storage.
// L - Liskov Substitution: Can replace any UserRepository implementation
type InMemoryUserRepository struct {
	users       map[string]*models.User
	usersByUser map[string]string // username -> id
	usersByEmail map[string]string // email -> id
	mu          sync.RWMutex
}

// InMemoryFriendshipRepository implements FriendshipRepository with thread-safe in-memory storage.
type InMemoryFriendshipRepository struct {
	friendships     map[string]*models.Friendship
	byRequesterRecv map[string]string // "requesterID:receiverID" -> id
	byUser          map[string][]string // userID -> []friendshipID
	mu              sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:        make(map[string]*models.User),
		usersByUser:  make(map[string]string),
		usersByEmail: make(map[string]string),
	}
}

// NewInMemoryFriendshipRepository creates a new in-memory friendship repository
func NewInMemoryFriendshipRepository() *InMemoryFriendshipRepository {
	return &InMemoryFriendshipRepository{
		friendships:     make(map[string]*models.Friendship),
		byRequesterRecv: make(map[string]string),
		byUser:          make(map[string][]string),
	}
}

// UserRepository implementation
func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.usersByUser[strings.ToLower(user.Username)]; exists {
		return ErrUserAlreadyExists
	}
	if _, exists := r.usersByEmail[strings.ToLower(user.Email)]; exists {
		return ErrUserAlreadyExists
	}

	r.users[user.ID] = user
	r.usersByUser[strings.ToLower(user.Username)] = user.ID
	r.usersByEmail[strings.ToLower(user.Email)] = user.ID
	return nil
}

func (r *InMemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usersByUser[strings.ToLower(username)]
	if !exists {
		return nil, ErrUserNotFound
	}
	return r.users[id], nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usersByEmail[strings.ToLower(email)]
	if !exists {
		return nil, ErrUserNotFound
	}
	return r.users[id], nil
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Search(query string, limit int) ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []*models.User{}, nil
	}

	var results []*models.User
	for _, user := range r.users {
		if !user.IsActive {
			continue
		}
		if strings.Contains(strings.ToLower(user.Username), query) ||
			strings.Contains(strings.ToLower(user.Email), query) ||
			strings.Contains(strings.ToLower(user.Bio), query) {
			results = append(results, user)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

// FriendshipRepository implementation
func (r *InMemoryFriendshipRepository) Create(friendship *models.Friendship) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := friendship.RequesterID + ":" + friendship.ReceiverID
	if _, exists := r.byRequesterRecv[key]; exists {
		return errors.New("friendship already exists")
	}

	r.friendships[friendship.ID] = friendship
	r.byRequesterRecv[key] = friendship.ID
	r.byUser[friendship.RequesterID] = append(r.byUser[friendship.RequesterID], friendship.ID)
	r.byUser[friendship.ReceiverID] = append(r.byUser[friendship.ReceiverID], friendship.ID)
	return nil
}

func (r *InMemoryFriendshipRepository) GetByID(id string) (*models.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, exists := r.friendships[id]
	if !exists {
		return nil, ErrFriendshipNotFound
	}
	return f, nil
}

func (r *InMemoryFriendshipRepository) GetPendingForUser(userID string) ([]*models.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Friendship
	for _, f := range r.friendships {
		if f.ReceiverID == userID && f.Status == models.FriendshipStatusPending {
			result = append(result, f)
		}
	}
	return result, nil
}

func (r *InMemoryFriendshipRepository) GetAcceptedFriends(userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	friendSet := make(map[string]bool)
	for _, f := range r.friendships {
		if f.Status != models.FriendshipStatusAccepted {
			continue
		}
		if f.RequesterID == userID {
			friendSet[f.ReceiverID] = true
		} else if f.ReceiverID == userID {
			friendSet[f.RequesterID] = true
		}
	}

	var friends []string
	for id := range friendSet {
		friends = append(friends, id)
	}
	return friends, nil
}

func (r *InMemoryFriendshipRepository) GetFollowers(userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var followers []string
	for _, f := range r.friendships {
		if f.Status == models.FriendshipStatusAccepted && f.ReceiverID == userID {
			followers = append(followers, f.RequesterID)
		}
	}
	return followers, nil
}

func (r *InMemoryFriendshipRepository) GetFollowing(userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var following []string
	for _, f := range r.friendships {
		if f.Status == models.FriendshipStatusAccepted && f.RequesterID == userID {
			following = append(following, f.ReceiverID)
		}
	}
	return following, nil
}

func (r *InMemoryFriendshipRepository) Update(friendship *models.Friendship) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.friendships[friendship.ID]; !exists {
		return ErrFriendshipNotFound
	}
	r.friendships[friendship.ID] = friendship
	return nil
}

func (r *InMemoryFriendshipRepository) Delete(requesterID, receiverID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := requesterID + ":" + receiverID
	id, exists := r.byRequesterRecv[key]
	if !exists {
		return ErrFriendshipNotFound
	}

	delete(r.friendships, id)
	delete(r.byRequesterRecv, key)
	return nil
}

func (r *InMemoryFriendshipRepository) GetFriendship(requesterID, receiverID string) (*models.Friendship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := requesterID + ":" + receiverID
	id, exists := r.byRequesterRecv[key]
	if !exists {
		return nil, ErrFriendshipNotFound
	}
	return r.friendships[id], nil
}
