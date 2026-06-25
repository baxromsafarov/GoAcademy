// Package gamification turns user activity into XP, levels and (CH11.2) streaks.
// It owns the XP policy and persists the user_stats aggregate transactionally
// alongside the activity that drives it.
package gamification

import "math"

// xpByActivity is the single source of truth for how much XP each activity type
// awards. An unknown type awards nothing.
var xpByActivity = map[string]int{
	"video_completed":           10,
	"article_read":              5,
	"quiz_passed":               20,
	"quiz_attempt":              2,
	"problem_solved":            30,
	"project_completed":         50,
	"daily_challenge_completed": 15, // CH11.4
}

// XPFor returns the XP awarded for an activity type (0 if unknown).
func XPFor(activityType string) int { return xpByActivity[activityType] }

// xpLevelBase scales the level curve: level L is reached at base*(L-1)^2 XP, so
// each level costs progressively more XP.
const xpLevelBase = 100

// LevelForXP maps lifetime XP to a level via level = 1 + floor(sqrt(xp/base)).
// 0 XP is level 1; with base=100: 100→L2, 400→L3, 900→L4.
func LevelForXP(totalXP int) int {
	if totalXP <= 0 {
		return 1
	}
	return 1 + int(math.Sqrt(float64(totalXP)/float64(xpLevelBase)))
}
