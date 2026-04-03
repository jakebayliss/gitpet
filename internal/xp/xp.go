package xp

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type CommitStats struct {
	LinesChanged int
	FilesTouched int
	Branch       string
	Repo         string
}

type XPResult struct {
	BaseXP       int
	FileBonus    int
	SizePenalty  float64
	BranchBonus  float64
	StreakBonus   float64
	TotalXP      int
	StreakDays   int
}

func ParseLastCommit() (*CommitStats, error) {
	statOut, err := exec.Command("git", "diff", "--stat", "HEAD~1", "HEAD").Output()
	if err != nil {
		statOut, err = exec.Command("git", "diff", "--stat", "--cached", "HEAD").Output()
		if err != nil {
			return &CommitStats{LinesChanged: 10, FilesTouched: 1}, nil
		}
	}

	lines := strings.Split(strings.TrimSpace(string(statOut)), "\n")
	if len(lines) == 0 {
		return &CommitStats{LinesChanged: 10, FilesTouched: 1}, nil
	}

	summary := lines[len(lines)-1]
	stats := &CommitStats{}

	parts := strings.Split(summary, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "file") {
			num := extractNumber(part)
			stats.FilesTouched = num
		} else if strings.Contains(part, "insertion") {
			stats.LinesChanged += extractNumber(part)
		} else if strings.Contains(part, "deletion") {
			stats.LinesChanged += extractNumber(part)
		}
	}

	branchOut, err := exec.Command("git", "branch", "--show-current").Output()
	if err == nil {
		stats.Branch = strings.TrimSpace(string(branchOut))
	}

	repoOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(repoOut)), "/")
		if len(parts) > 0 {
			stats.Repo = parts[len(parts)-1]
		}
	}

	if stats.LinesChanged == 0 {
		stats.LinesChanged = 10
	}
	if stats.FilesTouched == 0 {
		stats.FilesTouched = 1
	}

	return stats, nil
}

func Calculate(stats *CommitStats, streakDays int) *XPResult {
	result := &XPResult{
		SizePenalty: 1.0,
		BranchBonus: 1.0,
		StreakBonus:  1.0,
		StreakDays:   streakDays,
	}

	// Base XP: lines changed, capped at 200
	result.BaseXP = stats.LinesChanged
	if result.BaseXP > 200 {
		result.BaseXP = 200
	}

	// File bonus: 5 per file, capped at 10 files
	files := stats.FilesTouched
	if files > 10 {
		files = 10
	}
	result.FileBonus = files * 5

	// Size penalty for monster commits
	if stats.LinesChanged > 500 {
		result.SizePenalty = 0.5
	}

	// Branch bonus: not on main/master
	branch := stats.Branch
	if branch != "" && branch != "main" && branch != "master" {
		result.BranchBonus = 1.25
	}

	// Streak bonus: 1.1x per day after day 2, cap at 1.5x
	if streakDays >= 3 {
		bonus := 1.0 + float64(streakDays-2)*0.1
		if bonus > 1.5 {
			bonus = 1.5
		}
		result.StreakBonus = bonus
	}

	raw := float64(result.BaseXP+result.FileBonus) * result.SizePenalty * result.BranchBonus * result.StreakBonus
	result.TotalXP = int(math.Round(raw))
	if result.TotalXP < 1 {
		result.TotalXP = 1
	}

	return result
}

// XPForLevel returns total XP needed to reach a given level.
// Uses a RuneScape-inspired curve.
func XPForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	total := 0
	for l := 2; l <= level; l++ {
		total += int(float64(l-1) * 100 * math.Pow(1.08, float64(l-1)))
	}
	return total
}

func LevelFromXP(totalXP int) int {
	level := 1
	for level < 99 {
		if totalXP < XPForLevel(level+1) {
			break
		}
		level++
	}
	return level
}

func XPProgressString(xp, level int) string {
	currentLevelXP := XPForLevel(level)
	nextLevelXP := XPForLevel(level + 1)
	if level >= 99 {
		return fmt.Sprintf("MAX")
	}
	progress := xp - currentLevelXP
	needed := nextLevelXP - currentLevelXP
	return fmt.Sprintf("%d/%d", progress, needed)
}

func XPProgressBar(xp, level, width int) string {
	if level >= 99 {
		return strings.Repeat("=", width)
	}
	currentLevelXP := XPForLevel(level)
	nextLevelXP := XPForLevel(level + 1)
	progress := xp - currentLevelXP
	needed := nextLevelXP - currentLevelXP
	if needed <= 0 {
		return strings.Repeat("=", width)
	}
	filled := int(float64(progress) / float64(needed) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
}

func extractNumber(s string) int {
	var numStr string
	for _, c := range s {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		}
	}
	n, _ := strconv.Atoi(numStr)
	return n
}
