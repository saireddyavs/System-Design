package services

import (
	"buzzer-game/internal/interfaces"
	"buzzer-game/internal/models"
	"fmt"
	"sync"
	"time"
)

// GameService orchestrates the entire quiz game lifecycle.
// It coordinates players, rounds, questions, buzzer resolution,
// scoring, and leaderboard generation.
type GameService struct {
	playerRepo   interfaces.PlayerRepository
	quizRepo     interfaces.QuizRepository
	buzzerSvc    *BuzzerService
	leaderGen    interfaces.LeaderboardGenerator

	mu           sync.RWMutex
	state        models.GameState
	activeQuiz   *models.Quiz
	currentRound int
	currentQIdx  int
	answerLog    []models.AnswerEvent
	leaderboards []*models.Leaderboard
}

func NewGameService(
	playerRepo interfaces.PlayerRepository,
	quizRepo interfaces.QuizRepository,
	buzzerSvc *BuzzerService,
	leaderGen interfaces.LeaderboardGenerator,
) *GameService {
	return &GameService{
		playerRepo: playerRepo,
		quizRepo:   quizRepo,
		buzzerSvc:  buzzerSvc,
		leaderGen:  leaderGen,
		state:      models.GameStateWaiting,
	}
}

// RegisterPlayer adds a player to the game.
func (gs *GameService) RegisterPlayer(name string) (*models.Player, error) {
	player := models.NewPlayer(name)
	if err := gs.playerRepo.Add(player); err != nil {
		return nil, err
	}
	return player, nil
}

// CreateQuiz builds and stores a quiz.
func (gs *GameService) CreateQuiz(title string, rounds []*models.Round) (*models.Quiz, error) {
	quiz := models.NewQuiz(title, rounds)
	if err := gs.quizRepo.Save(quiz); err != nil {
		return nil, err
	}
	return quiz, nil
}

// StartGame begins the quiz game with the given quiz ID.
func (gs *GameService) StartGame(quizID string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.state == models.GameStateInProgress {
		return models.ErrGameAlreadyActive
	}

	quiz, err := gs.quizRepo.GetByID(quizID)
	if err != nil {
		return err
	}

	gs.activeQuiz = quiz
	gs.currentRound = 0
	gs.currentQIdx = 0
	gs.state = models.GameStateInProgress
	gs.answerLog = nil
	gs.leaderboards = nil

	return nil
}

// GetCurrentQuestion returns the active question.
func (gs *GameService) GetCurrentQuestion() (*models.Question, int, int, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.state != models.GameStateInProgress {
		return nil, 0, 0, models.ErrGameNotStarted
	}
	if gs.currentRound >= len(gs.activeQuiz.Rounds) {
		return nil, 0, 0, models.ErrGameNotStarted
	}

	round := gs.activeQuiz.Rounds[gs.currentRound]
	if gs.currentQIdx >= len(round.Questions) {
		return nil, 0, 0, models.ErrGameNotStarted
	}

	return round.Questions[gs.currentQIdx], gs.currentRound + 1, gs.currentQIdx + 1, nil
}

// StartQuestion opens the buzzer for the current question.
func (gs *GameService) StartQuestion() (*models.Question, error) {
	q, _, _, err := gs.GetCurrentQuestion()
	if err != nil {
		return nil, err
	}
	gs.buzzerSvc.OpenForQuestion(q.ID)
	return q, nil
}

// PressBuzzer allows a player to press the buzzer.
func (gs *GameService) PressBuzzer(playerID string) error {
	_, err := gs.playerRepo.GetByID(playerID)
	if err != nil {
		return err
	}
	return gs.buzzerSvc.PressBuzzer(playerID)
}

// ResolveBuzzer picks the buzzer winner and gives them the answer window.
func (gs *GameService) ResolveBuzzer() (string, error) {
	q, _, _, err := gs.GetCurrentQuestion()
	if err != nil {
		return "", err
	}
	return gs.buzzerSvc.ResolveBuzzer(q.AnswerTimeSec)
}

// SubmitAnswer processes an answer from the buzzer holder.
// Returns the AnswerEvent with the result and score delta.
func (gs *GameService) SubmitAnswer(playerID string, chosenOption string) (*models.AnswerEvent, error) {
	q, _, _, err := gs.GetCurrentQuestion()
	if err != nil {
		return nil, err
	}

	holder, withinTime := gs.buzzerSvc.GetBuzzerHolder()
	if holder != playerID {
		return nil, models.ErrNotBuzzerHolder
	}

	validOption := false
	for _, opt := range q.Options {
		if opt.Label == chosenOption {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, models.ErrInvalidOption
	}

	var result models.AnswerResult
	var delta int

	if !withinTime {
		result = models.AnswerResultTimeout
		delta = q.PenaltyTimeout
	} else if q.IsCorrect(chosenOption) {
		result = models.AnswerResultCorrect
		delta = q.PointsCorrect
	} else {
		result = models.AnswerResultIncorrect
		delta = q.PenaltyWrong
	}

	player, _ := gs.playerRepo.GetByID(playerID)
	player.AddScore(delta)

	event := &models.AnswerEvent{
		PlayerID:     playerID,
		QuestionID:   q.ID,
		ChosenOption: chosenOption,
		Result:       result,
		PointsDelta:  delta,
		AnsweredAt:   time.Now(),
	}

	gs.mu.Lock()
	gs.answerLog = append(gs.answerLog, *event)
	gs.mu.Unlock()

	gs.buzzerSvc.ClearBuzzerHolder()

	return event, nil
}

// HandleIncorrectOrTimeout is called after a wrong/timed-out answer.
// It excludes the player and reopens the buzzer for remaining players.
// Returns true if there are still eligible players, false if all are excluded.
func (gs *GameService) HandleIncorrectOrTimeout(playerID string) bool {
	gs.buzzerSvc.ExcludePlayer(playerID)
	gs.buzzerSvc.ReopenBuzzer()

	totalPlayers := len(gs.playerRepo.GetAll())
	excludedCount := gs.buzzerSvc.GetExcludedCount()

	return excludedCount < totalPlayers
}

// AdvanceQuestion moves to the next question. Returns false if the round is over.
func (gs *GameService) AdvanceQuestion() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	round := gs.activeQuiz.Rounds[gs.currentRound]
	gs.currentQIdx++
	return gs.currentQIdx < len(round.Questions)
}

// AdvanceRound moves to the next round. Generates a leaderboard for the
// completed round. Returns false if the game is over.
func (gs *GameService) AdvanceRound() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	roundNum := gs.currentRound + 1
	lb := gs.leaderGen.Generate(gs.playerRepo.GetAll(), roundNum)
	gs.leaderboards = append(gs.leaderboards, lb)

	gs.currentRound++
	gs.currentQIdx = 0

	if gs.currentRound >= len(gs.activeQuiz.Rounds) {
		gs.state = models.GameStateFinished
		return false
	}

	gs.state = models.GameStateInProgress
	return true
}

// GetLeaderboard returns the leaderboard for a given round (1-indexed).
func (gs *GameService) GetLeaderboard(roundNumber int) *models.Leaderboard {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for _, lb := range gs.leaderboards {
		if lb.RoundNumber == roundNumber {
			return lb
		}
	}
	return nil
}

// GetFinalLeaderboard returns the overall game leaderboard.
func (gs *GameService) GetFinalLeaderboard() *models.Leaderboard {
	return gs.leaderGen.Generate(gs.playerRepo.GetAll(), 0)
}

// GetState returns the current game state.
func (gs *GameService) GetState() models.GameState {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.state
}

// GetAnswerLog returns all answer events so far.
func (gs *GameService) GetAnswerLog() []models.AnswerEvent {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	log := make([]models.AnswerEvent, len(gs.answerLog))
	copy(log, gs.answerLog)
	return log
}

// PrintLeaderboard prints a formatted leaderboard.
func PrintLeaderboard(lb *models.Leaderboard) {
	if lb.RoundNumber == 0 {
		fmt.Println("\n=== FINAL LEADERBOARD ===")
	} else {
		fmt.Printf("\n=== LEADERBOARD (Round %d) ===\n", lb.RoundNumber)
	}
	fmt.Printf("%-6s %-20s %s\n", "Rank", "Player", "Score")
	fmt.Println("--------------------------------------")
	for _, e := range lb.Entries {
		fmt.Printf("%-6d %-20s %d\n", e.Rank, e.PlayerName, e.Score)
	}
}
