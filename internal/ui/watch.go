package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jakebayliss/gitpet/internal/evolution"
	"github.com/jakebayliss/gitpet/internal/store"
	"github.com/jakebayliss/gitpet/internal/xp"
)

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	rarityColors = map[string]lipgloss.Color{
		"Common":    lipgloss.Color("252"),
		"Rare":      lipgloss.Color("39"),
		"Legendary": lipgloss.Color("220"),
	}

	shinyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226"))

	levelUpStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46"))

	dropStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208"))

	statNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("105"))

	statBarFull = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	statBarEmpty = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type tickMsg time.Time

type WatchModel struct {
	store     *store.Store
	pet       *store.Pet
	hasStone  bool
	lastXP    int
	lastLevel int
	event     string
	eventTime time.Time
	spinner   spinner.Model
	width     int
	height    int
	err       error
}

func NewWatchModel(s *store.Store) WatchModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return WatchModel{
		store:   s,
		spinner: sp,
	}
}

func (m WatchModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		pollTick(),
	)
}

func pollTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m WatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		pet, err := m.store.GetActivePet()
		if err != nil {
			m.err = err
			return m, pollTick()
		}

		hasStone, _ := m.store.HasItem("evolution_stone")
		m.hasStone = hasStone

		// Detect changes
		if m.pet != nil {
			if pet.XP > m.lastXP && pet.Level > m.lastLevel {
				m.event = fmt.Sprintf("LEVEL UP! %s is now Level %d!", pet.Name, pet.Level)
				m.eventTime = time.Now()
			} else if pet.XP > m.lastXP {
				gained := pet.XP - m.lastXP
				m.event = fmt.Sprintf("+%d XP!", gained)
				m.eventTime = time.Now()
			}
		}

		m.lastXP = pet.XP
		m.lastLevel = pet.Level
		m.pet = pet

		return m, pollTick()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m WatchModel) View() string {
	if m.pet == nil {
		return "\n  Loading...\n"
	}

	p := m.pet
	var b strings.Builder

	// Header
	rarityColor, ok := rarityColors[p.Rarity]
	if !ok {
		rarityColor = lipgloss.Color("252")
	}
	rarityStyle := lipgloss.NewStyle().Foreground(rarityColor)

	name := titleStyle.Render(p.Name)
	species := p.Species
	rarity := rarityStyle.Render(p.Rarity)

	header := fmt.Sprintf("%s  -  %s  -  %s", name, species, rarity)

	if p.Prestige > 0 {
		stars := strings.Repeat("*", p.Prestige)
		header += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(stars)
	}

	if p.Shiny {
		header += "  " + shinyStyle.Render("SHINY")
	}

	b.WriteString(header + "\n")

	if p.Evolution > 1 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("Stage %d", p.Evolution)) + "\n")
	}

	// Art
	art := GetArtStage(p.Species, p.Evolution)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Render(art))
	b.WriteString("\n\n")

	// Level bar
	progressBar := renderColorBar(p.XP, p.Level, 20)
	progressStr := xp.XPProgressString(p.XP, p.Level)
	b.WriteString(fmt.Sprintf("  LVL %-3d %s  %s\n", p.Level, progressBar, dimStyle.Render(progressStr)))

	b.WriteString("\n")

	// Stats
	evo := p.Evolution
	renderStat := func(name string, val int) string {
		effective := evolution.EffectiveStat(val, evo)
		barWidth := 10
		filled := effective * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := statBarFull.Render(strings.Repeat("=", filled)) +
			statBarEmpty.Render(strings.Repeat("-", barWidth-filled))
		return fmt.Sprintf("  %s %s  %d", statNameStyle.Render(fmt.Sprintf("%-8s", name)), bar, effective)
	}

	b.WriteString(renderStat("WIT", p.Stats.Wit) + "\n")
	b.WriteString(renderStat("DEPTH", p.Stats.Depth) + "\n")
	b.WriteString(renderStat("STAMINA", p.Stats.Stamina) + "\n")
	b.WriteString(renderStat("LUCK", p.Stats.Luck) + "\n")
	b.WriteString(renderStat("ATTUNE", p.Stats.Attune) + "\n")

	b.WriteString("\n")

	// Evolution info
	if p.Evolution < 4 {
		nextLevel := evolution.NextStageLevel(p.Evolution)
		stoneText := dimStyle.Render("no")
		if m.hasStone {
			stoneText = dropStyle.Render("YES")
		}
		evoLine := fmt.Sprintf("  Next evo: Lv.%d  Stone: %s", nextLevel, stoneText)
		if p.Level >= nextLevel && m.hasStone {
			evoLine += "  " + levelUpStyle.Render("READY!")
		}
		b.WriteString(evoLine + "\n")
	} else {
		b.WriteString("  " + levelUpStyle.Render("MAX EVOLUTION") + "\n")
	}

	// Event banner
	if m.event != "" && time.Since(m.eventTime) < 10*time.Second {
		b.WriteString("\n")
		if strings.Contains(m.event, "LEVEL UP") {
			b.WriteString("  " + levelUpStyle.Render(m.event) + "\n")
		} else {
			b.WriteString("  " + dropStyle.Render(m.event) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + m.spinner.View() + " watching  |  q to quit"))

	content := borderStyle.Render(b.String())
	return "\n" + content + "\n"
}

func renderColorBar(currentXP, level, width int) string {
	if level >= 99 {
		return statBarFull.Render(strings.Repeat("=", width))
	}
	currentLevelXP := xp.XPForLevel(level)
	nextLevelXP := xp.XPForLevel(level + 1)
	progress := currentXP - currentLevelXP
	needed := nextLevelXP - currentLevelXP
	if needed <= 0 {
		return statBarFull.Render(strings.Repeat("=", width))
	}
	filled := progress * width / needed
	if filled > width {
		filled = width
	}
	return statBarFull.Render(strings.Repeat("=", filled)) +
		statBarEmpty.Render(strings.Repeat("-", width-filled))
}
