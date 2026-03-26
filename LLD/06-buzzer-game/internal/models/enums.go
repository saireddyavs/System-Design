package models

// GameState represents the current state of the game.
type GameState int

const (
	GameStateWaiting GameState = iota
	GameStateInProgress
	GameStateRoundComplete
	GameStateFinished
)

func (g GameState) String() string {
	switch g {
	case GameStateWaiting:
		return "WAITING"
	case GameStateInProgress:
		return "IN_PROGRESS"
	case GameStateRoundComplete:
		return "ROUND_COMPLETE"
	case GameStateFinished:
		return "FINISHED"
	default:
		return "UNKNOWN"
	}
}

// QuestionState represents the lifecycle of a single question.
type QuestionState int

const (
	QuestionStateOpen QuestionState = iota
	QuestionStateBuzzerPressed
	QuestionStateAnswered
	QuestionStateTimedOut
)

func (q QuestionState) String() string {
	switch q {
	case QuestionStateOpen:
		return "OPEN"
	case QuestionStateBuzzerPressed:
		return "BUZZER_PRESSED"
	case QuestionStateAnswered:
		return "ANSWERED"
	case QuestionStateTimedOut:
		return "TIMED_OUT"
	default:
		return "UNKNOWN"
	}
}

// AnswerResult represents the outcome of an answer attempt.
type AnswerResult int

const (
	AnswerResultCorrect AnswerResult = iota
	AnswerResultIncorrect
	AnswerResultTimeout
)

func (a AnswerResult) String() string {
	switch a {
	case AnswerResultCorrect:
		return "CORRECT"
	case AnswerResultIncorrect:
		return "INCORRECT"
	case AnswerResultTimeout:
		return "TIMEOUT"
	default:
		return "UNKNOWN"
	}
}
