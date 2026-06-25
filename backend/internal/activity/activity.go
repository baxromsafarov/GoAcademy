// Package activity defines the activity-recording seam used across the app to
// log meaningful user actions (which later drive the heatmap, XP and streaks).
//
// The seam (Recorder + Event) is consumed by the domain services; DBRecorder
// persists events to activity_log, while NoopRecorder is the inert default.
package activity

import "context"

// Event is a single recorded user action.
type Event struct {
	UserID  string // user UUID
	Type    string // e.g. "video_completed"
	RefType string // e.g. "video"
	RefID   string // referenced entity UUID
	XP      int    // XP awarded (0 until gamification, CHAPTER 11)
}

// Recorder records activity events. Implementations must be safe for concurrent use.
type Recorder interface {
	Record(ctx context.Context, ev Event) error
}

// NoopRecorder discards events. It is the inert default for contexts that do not
// persist activity (e.g. tests, or services wired before the DB is available).
type NoopRecorder struct{}

func (NoopRecorder) Record(context.Context, Event) error { return nil }
