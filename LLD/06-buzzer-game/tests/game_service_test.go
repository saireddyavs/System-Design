package tests

import (
	"buzzer-game/internal/interfaces"
	"buzzer-game/internal/models"
	"buzzer-game/internal/repositories"
	"buzzer-game/internal/services"
	"buzzer-game/internal/strategies"
	"testing"
)

func setupGame(t *testing.T) (*services.GameService, interfaces.PlayerRepository, *models.Player, *models.Player) {
	t.Helper()

	playerRepo := repositories.NewInMemoryPlayerRepository()
	quizRepo := repositories.NewInMemoryQuizRepository()
	buzzerSvc := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	leaderGen := strategies.NewScoreBasedLeaderboardGenerator()
	gameSvc := services.NewGameService(playerRepo, quizRepo, buzzerSvc, leaderGen)

	p1, _ := gameSvc.RegisterPlayer("Alice")
	p2, _ := gameSvc.RegisterPlayer("Bob")

	q1 := models.NewQuestion("What is 2+2?", []models.Option{
		{Label: "A", Text: "3"},
		{Label: "B", Text: "4"},
		{Label: "C", Text: "5"},
		{Label: "D", Text: "6"},
	}, "B")

	q2 := models.NewQuestion("What is 3+3?", []models.Option{
		{Label: "A", Text: "5"},
		{Label: "B", Text: "6"},
		{Label: "C", Text: "7"},
		{Label: "D", Text: "8"},
	}, "B")

	round := models.NewRound(1, []*models.Question{q1, q2})
	quiz, _ := gameSvc.CreateQuiz("Test Quiz", []*models.Round{round})
	_ = gameSvc.StartGame(quiz.ID)

	return gameSvc, playerRepo, p1, p2
}

func TestCorrectAnswer_Awards3Points(t *testing.T) {
	gameSvc, _, p1, _ := setupGame(t)

	q, _ := gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	event, err := gameSvc.SubmitAnswer(p1.ID, q.CorrectOption)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Result != models.AnswerResultCorrect {
		t.Errorf("expected CORRECT, got %s", event.Result)
	}
	if event.PointsDelta != 3 {
		t.Errorf("expected +3 points, got %d", event.PointsDelta)
	}
	if p1.GetScore() != 3 {
		t.Errorf("expected score 3, got %d", p1.GetScore())
	}
}

func TestIncorrectAnswer_Deducts1Point(t *testing.T) {
	gameSvc, _, p1, _ := setupGame(t)

	gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	event, err := gameSvc.SubmitAnswer(p1.ID, "D") // wrong
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Result != models.AnswerResultIncorrect {
		t.Errorf("expected INCORRECT, got %s", event.Result)
	}
	if event.PointsDelta != -1 {
		t.Errorf("expected -1 points, got %d", event.PointsDelta)
	}
}

func TestIncorrectAnswer_OpensForOtherPlayers(t *testing.T) {
	gameSvc, _, p1, p2 := setupGame(t)

	q, _ := gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	gameSvc.SubmitAnswer(p1.ID, "D") // wrong
	hasMore := gameSvc.HandleIncorrectOrTimeout(p1.ID)
	if !hasMore {
		t.Error("expected more eligible players")
	}

	_ = gameSvc.PressBuzzer(p2.ID)
	_, _ = gameSvc.ResolveBuzzer()

	event, err := gameSvc.SubmitAnswer(p2.ID, q.CorrectOption)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Result != models.AnswerResultCorrect {
		t.Errorf("expected CORRECT, got %s", event.Result)
	}
	if p2.GetScore() != 3 {
		t.Errorf("expected Bob score 3, got %d", p2.GetScore())
	}
}

func TestExcludedPlayer_CannotBuzz(t *testing.T) {
	gameSvc, _, p1, _ := setupGame(t)

	gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	gameSvc.SubmitAnswer(p1.ID, "D") // wrong
	gameSvc.HandleIncorrectOrTimeout(p1.ID)

	err := gameSvc.PressBuzzer(p1.ID)
	if err != models.ErrBuzzerLocked {
		t.Errorf("expected ErrBuzzerLocked, got %v", err)
	}
}

func TestNonHolder_CannotSubmitAnswer(t *testing.T) {
	gameSvc, _, p1, p2 := setupGame(t)

	gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	_, err := gameSvc.SubmitAnswer(p2.ID, "B")
	if err != models.ErrNotBuzzerHolder {
		t.Errorf("expected ErrNotBuzzerHolder, got %v", err)
	}
}

func TestLeaderboard_RankedByScore(t *testing.T) {
	gameSvc, _, p1, p2 := setupGame(t)

	// P1 answers Q1 correctly (+3)
	q1, _ := gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()
	gameSvc.SubmitAnswer(p1.ID, q1.CorrectOption)
	gameSvc.AdvanceQuestion()

	// P2 answers Q2 correctly (+3), P1 was wrong (-1)
	q2, _ := gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()
	gameSvc.SubmitAnswer(p1.ID, "D") // wrong: -1
	gameSvc.HandleIncorrectOrTimeout(p1.ID)

	_ = gameSvc.PressBuzzer(p2.ID)
	_, _ = gameSvc.ResolveBuzzer()
	gameSvc.SubmitAnswer(p2.ID, q2.CorrectOption) // correct: +3
	gameSvc.AdvanceQuestion()

	gameSvc.AdvanceRound()

	lb := gameSvc.GetLeaderboard(1)
	if lb == nil {
		t.Fatal("expected leaderboard, got nil")
	}
	if len(lb.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lb.Entries))
	}

	// Both have 3 and 2 points respectively, but Bob=3, Alice=2
	if lb.Entries[0].PlayerName != "Bob" {
		t.Errorf("expected Bob at rank 1, got %s", lb.Entries[0].PlayerName)
	}
	if lb.Entries[0].Score != 3 {
		t.Errorf("expected Bob score 3, got %d", lb.Entries[0].Score)
	}
	if lb.Entries[1].PlayerName != "Alice" {
		t.Errorf("expected Alice at rank 2, got %s", lb.Entries[1].PlayerName)
	}
	if lb.Entries[1].Score != 2 {
		t.Errorf("expected Alice score 2, got %d", lb.Entries[1].Score)
	}
}

func TestInvalidOption_ReturnsError(t *testing.T) {
	gameSvc, _, p1, _ := setupGame(t)

	gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()

	_, err := gameSvc.SubmitAnswer(p1.ID, "Z") // invalid
	if err != models.ErrInvalidOption {
		t.Errorf("expected ErrInvalidOption, got %v", err)
	}
}

func TestGameState_Transitions(t *testing.T) {
	gameSvc, _, _, _ := setupGame(t)

	if gameSvc.GetState() != models.GameStateInProgress {
		t.Errorf("expected IN_PROGRESS, got %s", gameSvc.GetState())
	}
}

func TestAllPlayersExcluded_ReturnsFalse(t *testing.T) {
	gameSvc, _, p1, p2 := setupGame(t)

	gameSvc.StartQuestion()
	_ = gameSvc.PressBuzzer(p1.ID)
	_, _ = gameSvc.ResolveBuzzer()
	gameSvc.SubmitAnswer(p1.ID, "D") // wrong
	gameSvc.HandleIncorrectOrTimeout(p1.ID)

	_ = gameSvc.PressBuzzer(p2.ID)
	_, _ = gameSvc.ResolveBuzzer()
	gameSvc.SubmitAnswer(p2.ID, "D") // wrong
	hasMore := gameSvc.HandleIncorrectOrTimeout(p2.ID)

	if hasMore {
		t.Error("expected no more eligible players")
	}
}
