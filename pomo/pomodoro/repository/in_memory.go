package repository

import (
	"fmt"
	"sync"

	"mnishiguchi.com/pomo/pomodoro"
)

// InMemoryRepo represents our in-memory repository.
type inMemoryRepo struct {
	// Because slices are not concurrent-safe, we use the mutex lock to prevent
	// concurrent-access to the data structure while making changes to it.
	sync.RWMutex
	intervals []pomodoro.Interval
}

func NewInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{
		intervals: []pomodoro.Interval{},
	}
}

// Create inserts a new entry for a provided pomodoro.Interval instance.
func (r *inMemoryRepo) Create(intvl pomodoro.Interval) (int64, error) {
	r.Lock() // Lock the writing
	defer r.Unlock()

	// Determine the ID for the new instance based on current record count.
	intvl.ID = int64(len(r.intervals)) + 1

	// Append the instance to the repo's internal list.
	r.intervals = append(r.intervals, intvl)

	return intvl.ID, nil
}

// Update updates the values of an existing entry in the repository.
func (r *inMemoryRepo) Update(intvl pomodoro.Interval) error {
	r.Lock() // Lock the writing
	defer r.Unlock()

	// ID is one based indexing so it must be positive interger.
	if intvl.ID < 1 {
		return fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, intvl.ID)
	}

	// The list index is zero indexed on the other hand.
	r.intervals[intvl.ID-1] = intvl

	return nil
}

// ByID finds an entry by ID.
func (r *inMemoryRepo) ByID(id int64) (pomodoro.Interval, error) {
	r.RLock() // Lock the reading
	defer r.RUnlock()

	intvl := pomodoro.Interval{}

	// ID is one based indexing so it must be positive interger.
	if id < 1 {
		return intvl, fmt.Errorf("%w: %d", pomodoro.ErrInvalidID, id)
	}

	// Find an entry by ID.
	intvl = r.intervals[id-1]

	return intvl, nil
}

// Last retrieves the last entry.
func (r *inMemoryRepo) Last() (pomodoro.Interval, error) {
	r.RLock() // Lock the reading
	defer r.RUnlock()

	intvl := pomodoro.Interval{}

	// Error is thre is no entry in the repository.
	if len(r.intervals) == 0 {
		return intvl, pomodoro.ErrNoIntervals
	}

	return r.intervals[len(r.intervals)-1], nil
}

// Breaks retrieves last N entries that are categorized as a break.
func (r *inMemoryRepo) Breaks(limit int) ([]pomodoro.Interval, error) {
	r.RLock() // Lock the reading
	defer r.RUnlock()

	data := []pomodoro.Interval{}

	// Loop over the data starting from the last item.
	for k := len(r.intervals) - 1; k >= 0; k-- {
		// Ignore CategoryPomodoro
		if r.intervals[k].Category == pomodoro.CategoryPomodoro {
			continue
		}

		// The other categories are breaks.
		data = append(data, r.intervals[k])

		// Halt if the limit is reached.
		if len(data) == limit {
			return data, nil
		}
	}

	return data, nil
}
