package main

import (
	"buzzer-game/internal/interfaces"
	"buzzer-game/internal/models"
	"buzzer-game/internal/repositories"
	"buzzer-game/internal/services"
	"buzzer-game/internal/strategies"
	"fmt"
	"time"
)

func main() {
	// --- Dependency wiring ---
	var playerRepo interfaces.PlayerRepository = repositories.NewInMemoryPlayerRepository()
	var quizRepo interfaces.QuizRepository = repositories.NewInMemoryQuizRepository()
	var buzzerStrategy interfaces.BuzzerStrategy = strategies.NewFCFSBuzzerStrategy()
	var leaderGen interfaces.LeaderboardGenerator = strategies.NewScoreBasedLeaderboardGenerator()

	buzzerSvc := services.NewBuzzerService(buzzerStrategy)
	gameSvc := services.NewGameService(playerRepo, quizRepo, buzzerSvc, leaderGen)

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║     MULTIPLAYER BUZZER QUIZ GAME         ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("Buzzer Strategy: %s\n\n", buzzerStrategy.Name())

	// --- Register players ---
	alice, _ := gameSvc.RegisterPlayer("Alice")
	bob, _ := gameSvc.RegisterPlayer("Bob")
	charlie, _ := gameSvc.RegisterPlayer("Charlie")
	dave, _ := gameSvc.RegisterPlayer("Dave")

	fmt.Println("Registered Players:")
	for _, p := range playerRepo.GetAll() {
		fmt.Printf("  - %s (ID: %s)\n", p.Name, p.ID[:8])
	}

	// --- Build quiz ---
	quiz := buildSampleQuiz()
	savedQuiz, _ := gameSvc.CreateQuiz(quiz.Title, quiz.Rounds)

	fmt.Printf("\nQuiz: %s\n", savedQuiz.Title)
	fmt.Printf("Rounds: %d, Total Questions: %d\n", len(savedQuiz.Rounds), countQuestions(savedQuiz))

	// --- Start game ---
	_ = gameSvc.StartGame(savedQuiz.ID)
	fmt.Println("\n>>> Game Started!")

	// =============================================
	// ROUND 1
	// =============================================
	fmt.Println("\n────────────────────────────────────")
	fmt.Println("  ROUND 1")
	fmt.Println("────────────────────────────────────")

	// Q1: Alice buzzes first and answers correctly (+3)
	simulateQuestion(gameSvc, 1, 1, func(q *models.Question) {
		fmt.Printf("  Alice buzzes first...\n")
		_ = gameSvc.PressBuzzer(alice.ID)
		time.Sleep(1 * time.Millisecond)
		_ = gameSvc.PressBuzzer(bob.ID) // too late
		winnerID, _ := gameSvc.ResolveBuzzer()
		printBuzzerWinner(playerRepo, winnerID)

		event, _ := gameSvc.SubmitAnswer(alice.ID, q.CorrectOption)
		printAnswerResult(playerRepo, event)
	})

	// Q2: Bob buzzes, answers wrong (-1), Charlie gets it (+3)
	simulateQuestion(gameSvc, 1, 2, func(q *models.Question) {
		fmt.Printf("  Bob buzzes first...\n")
		_ = gameSvc.PressBuzzer(bob.ID)
		winnerID, _ := gameSvc.ResolveBuzzer()
		printBuzzerWinner(playerRepo, winnerID)

		event, _ := gameSvc.SubmitAnswer(bob.ID, "D") // wrong
		printAnswerResult(playerRepo, event)

		hasMore := gameSvc.HandleIncorrectOrTimeout(bob.ID)
		fmt.Printf("  Remaining eligible players: %v\n", hasMore)

		fmt.Printf("  Charlie buzzes...\n")
		_ = gameSvc.PressBuzzer(charlie.ID)
		winnerID2, _ := gameSvc.ResolveBuzzer()
		printBuzzerWinner(playerRepo, winnerID2)

		event2, _ := gameSvc.SubmitAnswer(charlie.ID, q.CorrectOption)
		printAnswerResult(playerRepo, event2)
	})

	// Q3: Dave buzzes, times out (0 penalty), Alice answers correctly (+3)
	simulateQuestion(gameSvc, 1, 3, func(q *models.Question) {
		fmt.Printf("  Dave buzzes first...\n")
		_ = gameSvc.PressBuzzer(dave.ID)
		winnerID, _ := gameSvc.ResolveBuzzer()
		printBuzzerWinner(playerRepo, winnerID)

		// Simulate answer timeout: Dave answers but buzzer deadline passed
		event := &models.AnswerEvent{
			PlayerID:     dave.ID,
			QuestionID:   q.ID,
			ChosenOption: "",
			Result:       models.AnswerResultTimeout,
			PointsDelta:  q.PenaltyTimeout,
			AnsweredAt:   time.Now(),
		}
		dave.AddScore(q.PenaltyTimeout)
		fmt.Printf("  ⏱  %s timed out! (%+d points)\n", dave.Name, event.PointsDelta)

		hasMore := gameSvc.HandleIncorrectOrTimeout(dave.ID)
		fmt.Printf("  Remaining eligible players: %v\n", hasMore)

		fmt.Printf("  Alice buzzes...\n")
		_ = gameSvc.PressBuzzer(alice.ID)
		winnerID2, _ := gameSvc.ResolveBuzzer()
		printBuzzerWinner(playerRepo, winnerID2)

		event2, _ := gameSvc.SubmitAnswer(alice.ID, q.CorrectOption)
		printAnswerResult(playerRepo, event2)
	})

	gameSvc.AdvanceQuestion() // finish question movement
	moreRounds := gameSvc.AdvanceRound()
	services.PrintLeaderboard(gameSvc.GetLeaderboard(1))

	// =============================================
	// ROUND 2
	// =============================================
	if moreRounds {
		fmt.Println("\n────────────────────────────────────")
		fmt.Println("  ROUND 2")
		fmt.Println("────────────────────────────────────")

		// Q1: Charlie buzzes correctly (+3)
		simulateQuestion(gameSvc, 2, 1, func(q *models.Question) {
			fmt.Printf("  Charlie buzzes first...\n")
			_ = gameSvc.PressBuzzer(charlie.ID)
			winnerID, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID)

			event, _ := gameSvc.SubmitAnswer(charlie.ID, q.CorrectOption)
			printAnswerResult(playerRepo, event)
		})

		// Q2: Alice wrong (-1), Bob wrong (-1), Dave correct (+3)
		simulateQuestion(gameSvc, 2, 2, func(q *models.Question) {
			fmt.Printf("  Alice buzzes first...\n")
			_ = gameSvc.PressBuzzer(alice.ID)
			winnerID, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID)

			event, _ := gameSvc.SubmitAnswer(alice.ID, "A") // wrong
			printAnswerResult(playerRepo, event)

			gameSvc.HandleIncorrectOrTimeout(alice.ID)

			fmt.Printf("  Bob buzzes...\n")
			_ = gameSvc.PressBuzzer(bob.ID)
			winnerID2, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID2)

			event2, _ := gameSvc.SubmitAnswer(bob.ID, "A") // wrong too
			printAnswerResult(playerRepo, event2)

			gameSvc.HandleIncorrectOrTimeout(bob.ID)

			fmt.Printf("  Dave buzzes...\n")
			_ = gameSvc.PressBuzzer(dave.ID)
			winnerID3, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID3)

			event3, _ := gameSvc.SubmitAnswer(dave.ID, q.CorrectOption)
			printAnswerResult(playerRepo, event3)
		})

		// Q3: Nobody gets it right — all fail
		simulateQuestion(gameSvc, 2, 3, func(q *models.Question) {
			fmt.Printf("  Alice buzzes...\n")
			_ = gameSvc.PressBuzzer(alice.ID)
			winnerID, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID)

			event, _ := gameSvc.SubmitAnswer(alice.ID, "D") // wrong
			printAnswerResult(playerRepo, event)
			gameSvc.HandleIncorrectOrTimeout(alice.ID)

			fmt.Printf("  Bob buzzes...\n")
			_ = gameSvc.PressBuzzer(bob.ID)
			winnerID2, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID2)

			event2, _ := gameSvc.SubmitAnswer(bob.ID, "D") // wrong
			printAnswerResult(playerRepo, event2)
			gameSvc.HandleIncorrectOrTimeout(bob.ID)

			fmt.Printf("  Charlie buzzes...\n")
			_ = gameSvc.PressBuzzer(charlie.ID)
			winnerID3, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID3)

			event3, _ := gameSvc.SubmitAnswer(charlie.ID, "D") // wrong
			printAnswerResult(playerRepo, event3)
			gameSvc.HandleIncorrectOrTimeout(charlie.ID)

			fmt.Printf("  Dave buzzes...\n")
			_ = gameSvc.PressBuzzer(dave.ID)
			winnerID4, _ := gameSvc.ResolveBuzzer()
			printBuzzerWinner(playerRepo, winnerID4)

			event4, _ := gameSvc.SubmitAnswer(dave.ID, "D") // wrong
			printAnswerResult(playerRepo, event4)
			gameSvc.HandleIncorrectOrTimeout(dave.ID)

			fmt.Println("  No one answered correctly!")
		})

		gameSvc.AdvanceQuestion()
		gameSvc.AdvanceRound()
		services.PrintLeaderboard(gameSvc.GetLeaderboard(2))
	}

	// --- Final leaderboard ---
	services.PrintLeaderboard(gameSvc.GetFinalLeaderboard())
	fmt.Printf("\nGame State: %s\n", gameSvc.GetState())
	fmt.Println("\n>>> Game Over!")
}

func simulateQuestion(gs *services.GameService, round, qNum int, play func(q *models.Question)) {
	q, _ := gs.StartQuestion()
	fmt.Printf("\n  [Round %d, Q%d] %s\n", round, qNum, q.Text)
	for i, opt := range q.Options {
		fmt.Printf("    %s) %s", opt.Label, opt.Text)
		if i < len(q.Options)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	play(q)

	gs.AdvanceQuestion()
}

func printBuzzerWinner(repo interfaces.PlayerRepository, playerID string) {
	p, _ := repo.GetByID(playerID)
	fmt.Printf("  >> Buzzer won by: %s\n", p.Name)
}

func printAnswerResult(repo interfaces.PlayerRepository, event *models.AnswerEvent) {
	p, _ := repo.GetByID(event.PlayerID)
	switch event.Result {
	case models.AnswerResultCorrect:
		fmt.Printf("  ✓  %s answered '%s' — CORRECT! (%+d points, total: %d)\n",
			p.Name, event.ChosenOption, event.PointsDelta, p.GetScore())
	case models.AnswerResultIncorrect:
		fmt.Printf("  ✗  %s answered '%s' — WRONG! (%+d points, total: %d)\n",
			p.Name, event.ChosenOption, event.PointsDelta, p.GetScore())
	case models.AnswerResultTimeout:
		fmt.Printf("  ⏱  %s — TIMED OUT! (%+d points, total: %d)\n",
			p.Name, event.PointsDelta, p.GetScore())
	}
}

func buildSampleQuiz() *models.Quiz {
	// Round 1: 3 questions
	r1q1 := models.NewQuestion(
		"What is the time complexity of binary search?",
		[]models.Option{
			{Label: "A", Text: "O(n)"},
			{Label: "B", Text: "O(log n)"},
			{Label: "C", Text: "O(n log n)"},
			{Label: "D", Text: "O(1)"},
		},
		"B",
	)
	r1q2 := models.NewQuestion(
		"Which data structure uses LIFO?",
		[]models.Option{
			{Label: "A", Text: "Queue"},
			{Label: "B", Text: "Array"},
			{Label: "C", Text: "Stack"},
			{Label: "D", Text: "Linked List"},
		},
		"C",
	)
	r1q3 := models.NewQuestion(
		"What does CAP theorem stand for?",
		[]models.Option{
			{Label: "A", Text: "Cache, API, Protocol"},
			{Label: "B", Text: "Consistency, Availability, Partition Tolerance"},
			{Label: "C", Text: "Concurrency, Atomicity, Persistence"},
			{Label: "D", Text: "Compute, Access, Performance"},
		},
		"B",
	)

	// Round 2: 3 questions
	r2q1 := models.NewQuestion(
		"Which Go keyword is used for concurrency?",
		[]models.Option{
			{Label: "A", Text: "async"},
			{Label: "B", Text: "thread"},
			{Label: "C", Text: "go"},
			{Label: "D", Text: "spawn"},
		},
		"C",
	)
	r2q2 := models.NewQuestion(
		"What is the default port for HTTPS?",
		[]models.Option{
			{Label: "A", Text: "80"},
			{Label: "B", Text: "8080"},
			{Label: "C", Text: "443"},
			{Label: "D", Text: "3000"},
		},
		"C",
	)
	r2q3 := models.NewQuestion(
		"What is a mutex used for?",
		[]models.Option{
			{Label: "A", Text: "Sorting"},
			{Label: "B", Text: "Mutual exclusion"},
			{Label: "C", Text: "Memory allocation"},
			{Label: "D", Text: "Garbage collection"},
		},
		"B",
	)

	round1 := models.NewRound(1, []*models.Question{r1q1, r1q2, r1q3})
	round2 := models.NewRound(2, []*models.Question{r2q1, r2q2, r2q3})

	return &models.Quiz{
		Title:  "System Design & CS Fundamentals",
		Rounds: []*models.Round{round1, round2},
	}
}

func countQuestions(quiz *models.Quiz) int {
	total := 0
	for _, r := range quiz.Rounds {
		total += len(r.Questions)
	}
	return total
}
