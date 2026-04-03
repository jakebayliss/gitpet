package ui

import (
	"fmt"
	"strings"

	"github.com/petd/pet/internal/evolution"
	"github.com/petd/pet/internal/store"
	"github.com/petd/pet/internal/xp"
)

const cardWidth = 40

func RenderPetCard(p *store.Pet, hasStone bool) string {
	var b strings.Builder

	border := "+" + strings.Repeat("=", cardWidth-2) + "+"
	empty := "|" + strings.Repeat(" ", cardWidth-2) + "|"

	b.WriteString(border + "\n")

	// Name line with prestige stars
	stars := ""
	if p.Prestige > 0 {
		stars = " " + strings.Repeat("*", p.Prestige)
	}
	shinyMark := ""
	if p.Shiny {
		shinyMark = "  * SHINY"
	}
	nameLine := fmt.Sprintf("  %s  -  %s  -  %s%s", p.Name, p.Species, p.Rarity, stars)
	b.WriteString(padLine(nameLine))

	if p.Prestige > 0 {
		prestigeLine := fmt.Sprintf("  Prestige %d", p.Prestige)
		b.WriteString(padLine(prestigeLine))
	}

	b.WriteString(empty + "\n")

	// Evolution stage
	if p.Evolution > 1 {
		stageLine := fmt.Sprintf("  Stage %d", p.Evolution)
		b.WriteString(padLine(stageLine))
	}

	// ASCII art
	art := GetArtStage(p.Species, p.Evolution)
	artLines := strings.Split(strings.TrimPrefix(art, "\n"), "\n")
	for _, line := range artLines {
		if p.Shiny && line == artLines[len(artLines)/2] {
			// Add shiny marker on middle line
			combined := line + shinyMark
			b.WriteString(padLine(combined))
		} else {
			b.WriteString(padLine(line))
		}
	}

	b.WriteString(empty + "\n")

	// Level bar
	progressBar := xp.XPProgressBar(p.XP, p.Level, 12)
	progressStr := xp.XPProgressString(p.XP, p.Level)
	levelLine := fmt.Sprintf("  LVL %-3d %s  %s", p.Level, progressBar, progressStr)
	b.WriteString(padLine(levelLine))

	b.WriteString(empty + "\n")

	// Stats
	statBar := func(name string, val int) string {
		barWidth := 10
		filled := val * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)
		return fmt.Sprintf("  %-8s %s  %d", name, bar, val)
	}

	evo := p.Evolution
	b.WriteString(padLine(statBar("WIT", evolution.EffectiveStat(p.Stats.Wit, evo))))
	b.WriteString(padLine(statBar("DEPTH", evolution.EffectiveStat(p.Stats.Depth, evo))))
	b.WriteString(padLine(statBar("STAMINA", evolution.EffectiveStat(p.Stats.Stamina, evo))))
	b.WriteString(padLine(statBar("LUCK", evolution.EffectiveStat(p.Stats.Luck, evo))))
	b.WriteString(padLine(statBar("ATTUNE", evolution.EffectiveStat(p.Stats.Attune, evo))))

	b.WriteString(empty + "\n")

	// Evolution info
	if p.Evolution < 4 {
		nextLevel := evolution.NextStageLevel(p.Evolution)
		stoneStatus := "no"
		if hasStone {
			stoneStatus = "YES"
		}
		evolveReady := ""
		if p.Level >= nextLevel && hasStone {
			evolveReady = "  READY! run 'pet evolve'"
		}
		evoLine := fmt.Sprintf("  Next evo: Lv.%d  Stone: %s%s", nextLevel, stoneStatus, evolveReady)
		b.WriteString(padLine(evoLine))
	} else {
		b.WriteString(padLine("  MAX EVOLUTION"))
	}

	b.WriteString(border + "\n")

	return b.String()
}

func RenderHatch(p *store.Pet) string {
	var b strings.Builder

	b.WriteString("\nRolling...\n\n")

	shinyText := ""
	if p.Shiny {
		shinyText = " * SHINY!"
	}
	b.WriteString(fmt.Sprintf("You hatched a %s %s!%s\n", p.Rarity, p.Species, shinyText))
	b.WriteString(GetArt(p.Species) + "\n")

	return b.String()
}

func RenderCommitResult(petName string, xpGained int, totalXP int, level int, leveledUp bool, oldLevel int, streakDays int, branchBonus bool, dropItem string) string {
	var b strings.Builder

	progressStr := xp.XPProgressString(totalXP, level)

	if leveledUp {
		b.WriteString(fmt.Sprintf("\n*** LEVEL UP! %s is now Level %d! (was %d) ***\n", petName, level, oldLevel))
	}

	b.WriteString(fmt.Sprintf("%s gained %d XP! (%s)\n", petName, xpGained, progressStr))

	if branchBonus {
		b.WriteString("Branch bonus: 1.25x\n")
	}
	if streakDays >= 3 {
		bonus := 1.0 + float64(streakDays-2)*0.1
		if bonus > 1.5 {
			bonus = 1.5
		}
		b.WriteString(fmt.Sprintf("Streak: %d days (%.1fx bonus!)\n", streakDays, bonus))
	}

	if dropItem != "" && dropItem != "nothing" {
		switch dropItem {
		case "legendary_egg":
			b.WriteString("\n!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
			b.WriteString("!!!  LEGENDARY EGG DROPPED!  !!!\n")
			b.WriteString("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
			b.WriteString("Use 'pet spawn' to hatch it!\n")
		case "rare_egg":
			b.WriteString("\n*** RARE EGG DROPPED! ***\n")
			b.WriteString("Use 'pet spawn' to hatch it!\n")
		case "evolution_stone":
			b.WriteString("DROP: Evolution Stone!\n")
		case "xp_crystal":
			b.WriteString("DROP: XP Crystal! (+200 bonus XP)\n")
		case "title_scroll":
			b.WriteString("DROP: Title Scroll!\n")
		default:
			b.WriteString(fmt.Sprintf("DROP: %s!\n", dropItem))
		}
	}

	return b.String()
}

func padLine(content string) string {
	// Calculate visible length (content without escape sequences)
	visLen := len(content)
	if visLen >= cardWidth-2 {
		return "|" + content[:cardWidth-2] + "|\n"
	}
	padding := cardWidth - 2 - visLen
	return "|" + content + strings.Repeat(" ", padding) + "|\n"
}
