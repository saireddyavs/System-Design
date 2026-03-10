package repositories

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"sort"
	"strings"
	"sync"
)

// InMemoryChatRoomRepository implements ChatRoomRepository with in-memory storage
type InMemoryChatRoomRepository struct {
	rooms      map[string]*models.ChatRoom
	byUser     map[string][]string // userID -> room IDs
	oneOnOne   map[string]string   // "user1ID:user2ID" (sorted) -> room ID
	mu         sync.RWMutex
}

// NewInMemoryChatRoomRepository creates a new in-memory chat room repository
func NewInMemoryChatRoomRepository() interfaces.ChatRoomRepository {
	return &InMemoryChatRoomRepository{
		rooms:    make(map[string]*models.ChatRoom),
		byUser:   make(map[string][]string),
		oneOnOne: make(map[string]string),
	}
}

func oneOnOneKey(user1, user2 string) string {
	users := []string{user1, user2}
	sort.Strings(users)
	return strings.Join(users, ":")
}

func (r *InMemoryChatRoomRepository) Create(room *models.ChatRoom) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rooms[room.ID]; exists {
		return ErrAlreadyExists
	}

	roomCopy := *room
	membersCopy := make([]models.RoomMember, len(room.Members))
	copy(membersCopy, room.Members)
	roomCopy.Members = membersCopy

	r.rooms[room.ID] = &roomCopy

	for _, m := range room.Members {
		r.byUser[m.UserID] = append(r.byUser[m.UserID], room.ID)
	}

	if room.Type == models.ChatRoomTypeOneOnOne && len(room.Members) == 2 {
		key := oneOnOneKey(room.Members[0].UserID, room.Members[1].UserID)
		r.oneOnOne[key] = room.ID
	}

	return nil
}

func (r *InMemoryChatRoomRepository) GetByID(id string) (*models.ChatRoom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	room, exists := r.rooms[id]
	if !exists {
		return nil, ErrNotFound
	}
	roomCopy := *room
	membersCopy := make([]models.RoomMember, len(room.Members))
	copy(membersCopy, room.Members)
	roomCopy.Members = membersCopy
	return &roomCopy, nil
}

func (r *InMemoryChatRoomRepository) GetByUserID(userID string) ([]*models.ChatRoom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roomIDs, exists := r.byUser[userID]
	if !exists {
		return []*models.ChatRoom{}, nil
	}

	var rooms []*models.ChatRoom
	for _, roomID := range roomIDs {
		if room, ok := r.rooms[roomID]; ok {
			roomCopy := *room
			membersCopy := make([]models.RoomMember, len(room.Members))
			copy(membersCopy, room.Members)
			roomCopy.Members = membersCopy
			rooms = append(rooms, &roomCopy)
		}
	}
	return rooms, nil
}

func (r *InMemoryChatRoomRepository) Update(room *models.ChatRoom) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldRoom, exists := r.rooms[room.ID]
	if !exists {
		return ErrNotFound
	}

	// Rebuild byUser: remove old members, add new members
	oldMemberIDs := make(map[string]bool)
	for _, m := range oldRoom.Members {
		oldMemberIDs[m.UserID] = true
	}
	newMemberIDs := make(map[string]bool)
	for _, m := range room.Members {
		newMemberIDs[m.UserID] = true
	}

	// Remove room from users who left
	for uid := range oldMemberIDs {
		if !newMemberIDs[uid] {
			r.removeUserRoom(uid, room.ID)
		}
	}
	// Add room for new members
	for uid := range newMemberIDs {
		if !oldMemberIDs[uid] {
			r.byUser[uid] = append(r.byUser[uid], room.ID)
		}
	}

	roomCopy := *room
	membersCopy := make([]models.RoomMember, len(room.Members))
	copy(membersCopy, room.Members)
	roomCopy.Members = membersCopy
	r.rooms[room.ID] = &roomCopy
	return nil
}

func (r *InMemoryChatRoomRepository) removeUserRoom(userID, roomID string) {
	roomIDs := r.byUser[userID]
	var newIDs []string
	for _, id := range roomIDs {
		if id != roomID {
			newIDs = append(newIDs, id)
		}
	}
	if len(newIDs) == 0 {
		delete(r.byUser, userID)
	} else {
		r.byUser[userID] = newIDs
	}
}

func (r *InMemoryChatRoomRepository) GetOneOnOneRoom(user1ID, user2ID string) (*models.ChatRoom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := oneOnOneKey(user1ID, user2ID)
	roomID, exists := r.oneOnOne[key]
	if !exists {
		return nil, ErrNotFound
	}
	room := r.rooms[roomID]
	roomCopy := *room
	membersCopy := make([]models.RoomMember, len(room.Members))
	copy(membersCopy, room.Members)
	roomCopy.Members = membersCopy
	return &roomCopy, nil
}
