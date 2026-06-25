package gamification

import "time"

// streakUpdate is the recomputed streak after activity on a given UTC day.
type streakUpdate struct {
	Current    int
	Longest    int
	LastActive time.Time // UTC date
}

// computeStreak recomputes the streak after activity on newDate (a UTC date).
// prevActive is the previous last-active date; hasPrev is false when the user has
// never been active. Streaks count consecutive active days (D-010, UTC): the same
// day is unchanged, the next day increments, a gap of more than one day resets to
// 1, and an out-of-order older day is ignored. longest tracks the running maximum.
func computeStreak(prevActive time.Time, hasPrev bool, prevCurrent, prevLongest int, newDate time.Time) streakUpdate {
	if !hasPrev {
		return streakUpdate{Current: 1, Longest: max(prevLongest, 1), LastActive: newDate}
	}

	current := prevCurrent
	lastActive := prevActive
	switch {
	case !newDate.After(prevActive):
		// same day or a backfilled older day: nothing changes
	case newDate.Equal(prevActive.AddDate(0, 0, 1)):
		current = prevCurrent + 1
		lastActive = newDate
	default: // gap of more than one day: streak broken, restart
		current = 1
		lastActive = newDate
	}
	return streakUpdate{Current: current, Longest: max(prevLongest, current), LastActive: lastActive}
}
