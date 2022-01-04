package pomodoro

import (
	"context"
	"errors"
	"fmt"
	"time"
)

/*
The pomodoro technique records time in intervals that can be of different types
such as Pomodoro, short breaks, or long breaks.
*/

// Category constants
const (
	CategoryPomodoro   = "Pomodoro"
	CategoryShortBreak = "ShortBreak"
	CategoryLongBreak  = "LongBreak"
)

// State constants
const (
	// By using the iota operator, Go automatically increases the number for each line.
	StateNotStarted = iota
	StateRunning
	StatePaused
	StateDone
	StateCancelled
)

// Error values this package may return
var (
	ErrNoIntervals        = errors.New("No intervals")
	ErrIntervalNotRunning = errors.New("Interval not running")
	ErrIntervalCompleted  = errors.New("Interval is completed or cancelled")
	ErrInvalidState       = errors.New("Invalid State")
	ErrInvalidID          = errors.New("Invalid ID")
)

type Interval struct {
	ID              int64
	StartTime       time.Time
	PlannedDuration time.Duration
	ActualDuration  time.Duration
	Category        string
	State           int
}

// Repository abstracts the data source.
type Repository interface {
	Create(i Interval) (int64, error) // Creates an interval
	Update(i Interval) error          // Updates the interval
	ByID(id int64) (Interval, error)  // Retrieves an interval by ID
	Last() (Interval, error)          // Finds the last interval
	Breaks(n int) ([]Interval, error) // Retrieves intervals that matches CategoryLongBreak or CategoryShortBreak
}

// IntervalConfig represents the configuration required to instantiate an interval.
type IntervalConfig struct {
	repo               Repository
	PomodoroDuration   time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
}

// A callback function that is executed during the interval.
type Callback func(Interval)

// NewConfig instantiates a new IntervalConfig. The interval values falls back
// to default values when they are not provided.
func NewConfig(repo Repository, pomodoro, shortBreak, longBreak time.Duration) *IntervalConfig {
	c := &IntervalConfig{
		repo:               repo,
		PomodoroDuration:   25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
	}

	if pomodoro > 0 {
		c.PomodoroDuration = pomodoro
	}
	if shortBreak > 0 {
		c.ShortBreakDuration = shortBreak
	}
	if longBreak > 0 {
		c.LongBreakDuration = longBreak
	}

	return c
}

// Start starts the interval timer.
func (intvl Interval) Start(ctx context.Context, config *IntervalConfig, onStart, onTick, onEnd Callback) error {
	// Check the state of the current interval and set the appropriate options.
	switch intvl.State {
	// Do nothing if already running.
	case StateRunning:
		return nil

	// Start ticking if paused or not started.
	case StateNotStarted:
		intvl.StartTime = time.Now()
		fallthrough
	case StatePaused:
		// Mark as running, persist the record to the repo and start ticking.
		intvl.State = StateRunning
		if err := config.repo.Update(intvl); err != nil {
			return err
		}
		return tick(ctx, intvl.ID, config, onStart, onTick, onEnd)

	// Cannot start if cancelled or already done.
	case StateCancelled, StateDone:
		return fmt.Errorf("%w: Cannot start", ErrIntervalCompleted)

	// Invalid state
	default:
		return fmt.Errorf("%w: %d", ErrInvalidState, intvl.State)
	}
}

func (intvl Interval) Pause(config *IntervalConfig) error {
	// Cannot pause if not started.
	if intvl.State != StateRunning {
		return ErrIntervalNotRunning
	}

	// Mark as paused and persist the record to the repo.
	intvl.State = StatePaused
	return config.repo.Update(intvl)
}

// GetInterval finds the last interval or creates a new instance.
func GetInterval(config *IntervalConfig) (Interval, error) {
	intvl := Interval{}
	var err error

	// Get the last interval from the repo.
	intvl, err = config.repo.Last()

	// TODO: Maybe refactor later. These conditions can be better?
	// There is an error but it is not ErrNoIntervals. (???!!!)
	if err != nil &&
		err != ErrNoIntervals {
		return intvl, err
	}

	// No error + not cancelled + not done
	if err == nil &&
		intvl.State != StateCancelled &&
		intvl.State != StateDone {
		return intvl, nil
	}

	// If the last interval is inactive or unavailable, return a new interval.
	return newInterval(config)
}

func newInterval(config *IntervalConfig) (Interval, error) {
	intvl := Interval{}

	// Determine next category based on current state.
	category, err := nextCategory(config.repo)
	if err != nil {
		return intvl, err
	}
	intvl.Category = category

	// Determine the panned duration based on the category.
	switch category {
	case CategoryPomodoro:
		intvl.PlannedDuration = config.PomodoroDuration
	case CategoryShortBreak:
		intvl.PlannedDuration = config.ShortBreakDuration
	case CategoryLongBreak:
		intvl.PlannedDuration = config.LongBreakDuration
	}

	// Create a new record to the repo.
	if intvl.ID, err = config.repo.Create(intvl); err != nil {
		return intvl, err
	}

	return intvl, nil
}

// Tick controls the interval timer.
func tick(
	ctx context.Context, // Needed to receive signals
	id int64, // The ID of the interval control
	config *IntervalConfig, // An instance of the configuration
	onStart Callback, // A callback that fires when timer starts.
	onTick Callback, // A callback that fires every second.
	onExpire Callback, // A callback that fires when the timer expires.
) error {
	// Ticker allows us to execute actions every second in a loop.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Obtain the interval instance by ID.
	intvl, err := config.repo.ByID(id)
	if err != nil {
		return err
	}

	// Schedule to send notification when the planned duration has elapsed.
	expireCh := time.After(intvl.PlannedDuration - intvl.ActualDuration)

	onStart(intvl) // callback

	for {
		select {
		case <-ticker.C: // When time.Ticker goes off
			intvl, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			// Do nothing if paused
			if intvl.State == StatePaused {
				return nil
			}

			// Increment the actual duration and persist the record to the repo.
			intvl.ActualDuration += time.Second
			if err := config.repo.Update(intvl); err != nil {
				return err
			}

			onTick(intvl) // callback

		case <-expireCh: // When the interval time expires
			intvl, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			// Mark as done and persist the record to the repo.
			intvl.State = StateDone
			onExpire(intvl) // callback
			return config.repo.Update(intvl)

		case <-ctx.Done(): // When a signal is received from context
			intvl, err := config.repo.ByID(id)
			if err != nil {
				return err
			}

			// Mark as cancelled and persist the record to the repo.
			intvl.State = StateCancelled
			return config.repo.Update(intvl)
		}
	}
}

// NextCategory retrieves the next interval category from the repository and
// determines the text interval category based on the Pomodoro Technique rules.
//
// P1 -> short -> P2 -> short -> P3 -> short -> P4 -> long
//
// After each Pomodoro interval, there is a short break.
// After four Pomodoros, there is a long break.
func nextCategory(repo Repository) (string, error) {
	lastInterval, err := repo.Last()
	if err != nil {
		// For the first execution, return CategoryPomodoro.
		if err == ErrNoIntervals {
			return CategoryPomodoro, nil
		}

		return "", err
	}

	// After a break, always return CategoryPomodoro.
	if lastInterval.Category == CategoryLongBreak || lastInterval.Category == CategoryShortBreak {
		return CategoryPomodoro, nil
	}

	// Obtain last three breaks.
	lastBreaks, err := repo.Breaks(3)
	if err != nil {
		return "", err
	}

	// If the break record count is still less than three, return CategoryShortBreak.
	if len(lastBreaks) < 3 {
		return CategoryShortBreak, nil
	}

	// If there was CategoryShortBreak in the past three breaks, return CategoryShortBreak.
	for _, i := range lastBreaks {
		if i.Category == CategoryLongBreak {
			return CategoryShortBreak, nil
		}
	}

	// Finally we can take a long break.
	return CategoryLongBreak, nil
}
