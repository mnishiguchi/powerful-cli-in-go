package pomodoro_test

import (
	"testing"

	"mnishiguchi.com/pomo/pomodoro"
	"mnishiguchi.com/pomo/pomodoro/repository"
)

// GetRepo is a test helper that returns the repository instance and a cleanup function.
func getRepo(t *testing.T) (pomodoro.Repository, func()) {
	t.Helper()

	// The in-memory repository does not require a cleanup function.
	return repository.NewInMemoryRepo(), func() {}
}
