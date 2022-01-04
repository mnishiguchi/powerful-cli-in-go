package pomodoro_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"mnishiguchi.com/pomo/pomodoro"
)

func TestNewConfig(t *testing.T) {
	// Prepare test cases with a slice of annonymous structs.
	testCases := []struct {
		name           string
		durations      [3]time.Duration
		expectedConfig pomodoro.IntervalConfig
	}{
		{name: "default durations",
			durations: [3]time.Duration{},
			expectedConfig: pomodoro.IntervalConfig{
				PomodoroDuration:   25 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
			},
		},
		{name: "provide one duration",
			durations: [3]time.Duration{
				20 * time.Minute,
			},
			expectedConfig: pomodoro.IntervalConfig{
				PomodoroDuration:   20 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
			},
		},
		{name: "provide multiple durations",
			durations: [3]time.Duration{
				20 * time.Minute,
				10 * time.Minute,
				12 * time.Minute,
			},
			expectedConfig: pomodoro.IntervalConfig{
				PomodoroDuration:   20 * time.Minute,
				ShortBreakDuration: 10 * time.Minute,
				LongBreakDuration:  12 * time.Minute,
			},
		},
	}

	// Execute tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var repo pomodoro.Repository

			// Execute the subject
			actualConfig := pomodoro.NewConfig(repo, tc.durations[0], tc.durations[1], tc.durations[2])

			// Verify if the configuration matches the expected values.
			if actualConfig.PomodoroDuration != tc.expectedConfig.PomodoroDuration {
				t.Errorf("Expected Pomodoro Duration %q, got %q instead\n",
					tc.expectedConfig.PomodoroDuration, actualConfig.PomodoroDuration)
			}
			if actualConfig.ShortBreakDuration != tc.expectedConfig.ShortBreakDuration {
				t.Errorf("Expected ShortBreak Duration %q, got %q instead\n",
					tc.expectedConfig.ShortBreakDuration, actualConfig.ShortBreakDuration)
			}
			if actualConfig.LongBreakDuration != tc.expectedConfig.LongBreakDuration {
				t.Errorf("Expected LongBreak Duration %q, got %q instead\n",
					tc.expectedConfig.LongBreakDuration, actualConfig.LongBreakDuration)
			}
		})
	}
}

func TestGetInterval(t *testing.T) {
	repo, cleanup := getRepo(t)
	defer cleanup()

	// Create an interval configuration
	const duration = 1 * time.Millisecond
	config := pomodoro.NewConfig(repo, 3*duration, duration, 2*duration)

	// Execute the GetInterval() 16 times.
	for i := 1; i <= 16; i++ {
		var (
			expectedCategory string
			expectedDuration time.Duration
		)

		// Determine expecetd values for each iteration.
		// P1 -> short -> P2 -> short -> P3 -> short -> P4 -> long
		switch {
		case i%2 != 0:
			expectedCategory = pomodoro.CategoryPomodoro
			expectedDuration = 3 * duration
		case i%8 == 0:
			expectedCategory = pomodoro.CategoryLongBreak
			expectedDuration = 2 * duration
		case i%2 == 0:
			expectedCategory = pomodoro.CategoryShortBreak
			expectedDuration = duration
		}

		// Determine a test name based on iteration index and expected category.
		testName := fmt.Sprintf("%02d %s", i, expectedCategory)

		t.Run(testName, func(t *testing.T) {
			// Execute the subject.
			intvl, err := pomodoro.GetInterval(config)

			if err != nil {
				t.Errorf("Expected no error, got %q.\n", err)
			}

			// A callback function that does nothing.
			noop := func(pomodoro.Interval) {}

			// Start the interval timer.
			if err := intvl.Start(context.Background(), config, noop, noop, noop); err != nil {
				t.Fatal(err)
			}

			// Verify if the interval matches the expected values.
			if intvl.Category != expectedCategory {
				t.Errorf("Expected category %q, got %q.\n", expectedCategory, intvl.Category)
			}
			if intvl.PlannedDuration != expectedDuration {
				t.Errorf("Expected PlannedDuration %q, got %q.\n", expectedDuration, intvl.PlannedDuration)
			}
			if intvl.State != pomodoro.StateNotStarted {
				t.Errorf("Expected State = %q, got %q.\n", pomodoro.StateNotStarted, intvl.State)
			}

			// Reload the same record from the repository.
			intvlReloaded, err := repo.ByID(intvl.ID)
			if err != nil {
				t.Errorf("Expected no error. Got %q.\n", err)
			}

			// Verify if the reloaded record is marked as done.
			if intvlReloaded.State != pomodoro.StateDone {
				t.Errorf("Expected State = %q, got %q.\n", pomodoro.StateDone, intvlReloaded.State)
			}
		})
	}
}

func TestPause(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	// Create an interval configuration
	config := pomodoro.NewConfig(repo, duration, duration, duration)

	testCases := []struct {
		name             string
		shouldStart      bool
		expectedState    int
		expectedDuration time.Duration
	}{
		{name: "not started", shouldStart: false,
			expectedState:    pomodoro.StateNotStarted,
			expectedDuration: 0},
		{name: "paused", shouldStart: true,
			expectedState: pomodoro.StatePaused,
			// Should be one second because the time will be paused at the first tick.
			expectedDuration: time.Second},
	}

	expectedError := pomodoro.ErrIntervalNotRunning

	// Execute tests for Pause
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Get an interval record based on the configuration.
			intvl, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			// Determine callbacks.
			onStart := func(pomodoro.Interval) {}
			onExpire := func(pomodoro.Interval) {
				t.Errorf("End callback should not be executed")
			}
			// Pause the timer when ticked.
			onTick := func(i pomodoro.Interval) {
				if err := i.Pause(config); err != nil {
					t.Fatal(err)
				}
			}

			// Start the timer as needed.
			if tc.shouldStart {
				if err := intvl.Start(ctx, config, onStart, onTick, onExpire); err != nil {
					t.Fatal(err)
				}
			}

			// Get an interval record.
			intvl, err = pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			// Pause the timer and verify the expected error is raised.
			err = intvl.Pause(config)
			if err == nil {
				t.Errorf("Expected error %q, got nil", expectedError)
			} else {
				if !errors.Is(err, expectedError) {
					t.Fatalf("Expected error %q, got %q", expectedError, err)
				}
			}

			// Reload the record from the repository.
			intvl, err = repo.ByID(intvl.ID)
			if err != nil {
				t.Fatal(err)
			}

			// Verify the record matches the expectation.
			if intvl.State != tc.expectedState {
				t.Errorf("Expected state %d, got %d.\n", tc.expectedState, intvl.State)
			}
			if intvl.ActualDuration != tc.expectedDuration {
				t.Errorf("Expected duration %q, got %q.\n", tc.expectedDuration, intvl.ActualDuration)
			}

			cancel()
		})
	}
}

func TestStart(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	config := pomodoro.NewConfig(repo, duration, duration, duration)

	testCases := []struct {
		name             string
		shouldCancel     bool
		expectedState    int
		expectedDuration time.Duration
	}{
		{name: "finish", shouldCancel: false,
			expectedState:    pomodoro.StateDone,
			expectedDuration: duration},
		{name: "cancel", shouldCancel: true,
			expectedState:    pomodoro.StateCancelled,
			expectedDuration: time.Second},
	}

	// Execute tests for Start
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			intvl, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			// Callback functions
			// Note that the interval is passed in when the callback is executed.
			// Do not mix up with the instance in the outer scope.
			onStart := func(i pomodoro.Interval) {
				if i.State != pomodoro.StateRunning {
					t.Errorf("Expected state %d, got %d.\n", pomodoro.StateRunning, i.State)
				}
				if i.ActualDuration >= i.PlannedDuration {
					t.Errorf("Expected ActualDuration %q, less than Planned %q.\n", i.ActualDuration, i.PlannedDuration)
				}
			}
			onExpire := func(i pomodoro.Interval) {
				if i.State != tc.expectedState {
					t.Errorf("Expected state %d, got %d.\n", tc.expectedState, i.State)
				}
				if tc.shouldCancel {
					t.Errorf("OnExpire callback should not be executed")
				}
			}
			onTick := func(i pomodoro.Interval) {
				if i.State != pomodoro.StateRunning {
					t.Errorf("Expected state %d, got %d.\n", pomodoro.StateRunning, i.State)
				}
				if tc.shouldCancel {
					cancel()
				}
			}

			// Execute the Start method (subject)
			if err := intvl.Start(ctx, config, onStart, onTick, onExpire); err != nil {
				t.Fatal(err)
			}

			// Reload the interval record from the repository.
			intvl, err = repo.ByID(intvl.ID)
			if err != nil {
				t.Fatal(err)
			}

			if intvl.State != tc.expectedState {
				t.Errorf("Expected state %d, got %d.\n", tc.expectedState, intvl.State)
			}

			if intvl.ActualDuration != tc.expectedDuration {
				t.Errorf("Expected ActualDuration %q, got %q.\n", tc.expectedDuration, intvl.ActualDuration)
			}

			cancel()
		})
	}
}
