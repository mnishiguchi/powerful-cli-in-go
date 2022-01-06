package cmd

import (
	"mnishiguchi.com/pomo/pomodoro"
	"mnishiguchi.com/pomo/pomodoro/repository"
)

func getRepo() (pomodoro.Repository, error) {
	return repository.NewInMemoryRepo(), nil
}
