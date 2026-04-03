package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/petd/pet/internal/drop"
	"github.com/petd/pet/internal/evolution"
	"github.com/petd/pet/internal/gen"
	"github.com/petd/pet/internal/store"
	"github.com/petd/pet/internal/ui"
	"github.com/petd/pet/internal/xp"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "pet",
		Short: "Terminal pet companion — level up by committing code",
	}

	root.AddCommand(spawnCmd())
	root.AddCommand(showCmd())
	root.AddCommand(initCmd())
	root.AddCommand(commitCmd())
	root.AddCommand(inventoryCmd())
	root.AddCommand(evolveCmd())
	root.AddCommand(prestigeCmd())
	root.AddCommand(listCmd())
	root.AddCommand(switchCmd())
	root.AddCommand(killCmd())
	root.AddCommand(logCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func spawnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spawn",
		Short: "Spawn a new pet",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			hasPet, err := s.HasAnyPet()
			if err != nil {
				return err
			}

			var tier gen.Tier
			if !hasPet {
				// First pet is always common
				tier = gen.TierCommon
			} else {
				// Check for eggs in inventory
				hasRare, _ := s.HasItem("rare_egg")
				hasLegendary, _ := s.HasItem("legendary_egg")

				if hasLegendary {
					fmt.Print("You have a LEGENDARY EGG! Use it? [y/N] > ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(strings.ToLower(input))
					if input == "y" || input == "yes" {
						tier = gen.TierLegendary
						s.UseInventoryItem("legendary_egg")
					} else if hasRare {
						fmt.Print("Use your Rare Egg instead? [y/N] > ")
						input, _ = reader.ReadString('\n')
						input = strings.TrimSpace(strings.ToLower(input))
						if input == "y" || input == "yes" {
							tier = gen.TierRare
							s.UseInventoryItem("rare_egg")
						} else {
							fmt.Println("No egg used.")
							return nil
						}
					} else {
						fmt.Println("No egg used.")
						return nil
					}
				} else if hasRare {
					fmt.Print("You have a RARE EGG! Use it? [y/N] > ")
					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(strings.ToLower(input))
					if input == "y" || input == "yes" {
						tier = gen.TierRare
						s.UseInventoryItem("rare_egg")
					} else {
						fmt.Println("No egg used.")
						return nil
					}
				} else {
					fmt.Println("You need an egg to spawn another pet! Keep committing for drops.")
					return nil
				}
			}

			pet := gen.Roll(tier)

			fmt.Print(ui.RenderHatch(pet))

			fmt.Print("What will you name it? > ")
			reader := bufio.NewReader(os.Stdin)
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "" {
				name = "Unnamed"
			}
			pet.Name = name

			if err := s.CreatePet(pet); err != nil {
				return fmt.Errorf("saving pet: %w", err)
			}
			if err := s.SetActivePet(pet.ID); err != nil {
				return fmt.Errorf("setting active: %w", err)
			}

			fmt.Printf("\nWelcome, %s!\n", pet.Name)

			if !hasPet {
				// Auto-install global git hook on first pet only
				if err := installGlobalHook(); err != nil {
					fmt.Printf("Warning: couldn't install global git hook: %v\n", err)
					fmt.Println("You can try manually with 'pet init'")
				} else {
					fmt.Println("Global git hook installed — every commit in any repo feeds your pet!")
				}
			} else {
				fmt.Printf("%s is now your active pet.\n", pet.Name)
			}

			fmt.Println("\nRun 'pet show' to see your pet.")
			return nil
		},
	}
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show your active pet",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetActivePet()
			if err != nil {
				fmt.Println("No pet found! Run 'pet spawn' to get started.")
				return nil
			}

			hasStone, _ := s.HasItem("evolution_stone")
			fmt.Print(ui.RenderPetCard(pet, hasStone))
			return nil
		},
	}
}

func installGlobalHook() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}
	exePath = filepath.ToSlash(exePath)

	hookScript := fmt.Sprintf("#!/bin/sh\n\"%s\" commit\n", exePath)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	hooksDir := filepath.Join(home, ".petd", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}
	hookPath := filepath.Join(hooksDir, "post-commit")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return err
	}
	if err := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir).Run(); err != nil {
		return fmt.Errorf("setting global hooks path: %w", err)
	}
	return nil
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Reinstall global git hook (normally done automatically by pet hatch)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := installGlobalHook(); err != nil {
				return err
			}
			fmt.Println("Global git hook installed — every commit in any repo feeds your pet!")
			return nil
		},
	}
}

func commitCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "commit",
		Short:  "Record XP from last git commit (called by git hook)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetActivePet()
			if err != nil {
				return nil // silently exit if no pet
			}

			stats, err := xp.ParseLastCommit()
			if err != nil {
				return nil // silently exit on parse failure
			}

			streakDays, _ := s.GetStreakDays()
			result := xp.Calculate(stats, streakDays)

			// Roll for drop
			dropItem := drop.Roll()

			// Handle XP crystal immediately
			bonusXP := 0
			if dropItem == "xp_crystal" {
				bonusXP = 200
			}

			oldLevel := pet.Level
			pet.XP += result.TotalXP + bonusXP
			newLevel := xp.LevelFromXP(pet.XP)
			leveledUp := newLevel > oldLevel
			pet.Level = newLevel

			if err := s.UpdatePetXP(pet.ID, pet.XP, pet.Level); err != nil {
				return err
			}

			// Store drop in inventory
			if dropItem != "nothing" {
				if dropItem == "title_scroll" {
					// Store the actual title name
					if err := s.AddInventoryItem("title:"+drop.RandomTitle(), 1); err != nil {
						return err
					}
				} else if dropItem != "xp_crystal" {
					if err := s.AddInventoryItem(dropItem, 1); err != nil {
						return err
					}
				}
			}

			// Record commit
			commit := &store.Commit{
				PetID:        pet.ID,
				Repo:         stats.Repo,
				LinesChanged: stats.LinesChanged,
				FilesTouched: stats.FilesTouched,
				XPEarned:     result.TotalXP + bonusXP,
				DropItem:     dropItem,
				CreatedAt:    time.Now(),
			}
			if err := s.RecordCommit(commit); err != nil {
				return err
			}

			branchBonus := result.BranchBonus > 1.0
			fmt.Print(ui.RenderCommitResult(
				pet.Name, result.TotalXP+bonusXP, pet.XP, pet.Level,
				leveledUp, oldLevel, streakDays, branchBonus, dropItem,
			))

			return nil
		},
	}
}

func inventoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inventory",
		Short: "Show your collected items",
		Aliases: []string{"inv"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			items, err := s.GetInventory()
			if err != nil {
				return err
			}

			if len(items) == 0 {
				fmt.Println("Your inventory is empty. Keep committing!")
				return nil
			}

			fmt.Println("+========================+")
			fmt.Println("|      INVENTORY         |")
			fmt.Println("+========================+")
			for _, item := range items {
				display := drop.DisplayName(item.Item)
				if display == "" {
					display = item.Item
				}
				fmt.Printf("  %-25s x%d\n", display, item.Quantity)
			}
			fmt.Println()
			return nil
		},
	}
}

func evolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "evolve",
		Short: "Evolve your active pet to the next stage",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetActivePet()
			if err != nil {
				fmt.Println("No pet found! Run 'pet spawn' to get started.")
				return nil
			}

			if pet.Evolution >= 4 {
				fmt.Printf("%s is already at max evolution (Stage 4)!\n", pet.Name)
				return nil
			}

			hasStone, err := s.HasItem("evolution_stone")
			if err != nil {
				return err
			}

			nextLevel := evolution.NextStageLevel(pet.Evolution)

			if pet.Level < nextLevel {
				fmt.Printf("%s needs to be Level %d to evolve (current: %d)\n", pet.Name, nextLevel, pet.Level)
				return nil
			}

			if !hasStone {
				fmt.Printf("%s is ready to evolve but you need an Evolution Stone!\n", pet.Name)
				fmt.Println("Keep committing for a chance to drop one.")
				return nil
			}

			if err := s.UseInventoryItem("evolution_stone"); err != nil {
				return err
			}

			newStage := pet.Evolution + 1
			if err := s.UpdatePetEvolution(pet.ID, newStage); err != nil {
				return err
			}

			fmt.Printf("\n========================================\n")
			fmt.Printf("  %s is evolving...\n", pet.Name)
			fmt.Printf("  Stage %d -> Stage %d!\n", pet.Evolution, newStage)
			fmt.Printf("========================================\n")
			fmt.Println(ui.GetArtStage(pet.Species, newStage))
			fmt.Printf("  Stats multiplier: %.1fx\n", evolution.Stages[newStage-1].StatMultiplier)
			fmt.Printf("========================================\n\n")

			return nil
		},
	}
}

func prestigeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prestige",
		Short: "Reset to level 1 for a prestige star and permanent stat bonus (requires level 99)",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetActivePet()
			if err != nil {
				fmt.Println("No pet found! Run 'pet spawn' to get started.")
				return nil
			}

			if pet.Level < 99 {
				fmt.Printf("%s needs to be Level 99 to prestige (current: %d)\n", pet.Name, pet.Level)
				return nil
			}

			newPrestige := pet.Prestige + 1
			stars := strings.Repeat("*", newPrestige)

			fmt.Printf("\nAre you sure? %s will reset to Level 1.\n", pet.Name)
			fmt.Printf("  Prestige stars earned: %s\n", stars)
			fmt.Printf("  Permanent stat bonus: +2 to all stats\n")
			fmt.Printf("  Evolution resets to Stage 1\n")
			fmt.Print("\n[y/N] > ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input != "y" && input != "yes" {
				fmt.Println("Prestige cancelled.")
				return nil
			}

			if err := s.PrestigePet(pet.ID, newPrestige, 2); err != nil {
				return err
			}

			fmt.Printf("\n========================================\n")
			fmt.Printf("  %s has been reborn!\n", pet.Name)
			fmt.Printf("  %s Prestige %d %s\n", stars, newPrestige, stars)
			fmt.Printf("  +2 to all base stats permanently\n")
			fmt.Printf("========================================\n\n")

			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all your pets",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pets, err := s.GetAllPets()
			if err != nil {
				return err
			}

			if len(pets) == 0 {
				fmt.Println("No pets! Run 'pet spawn' to get started.")
				return nil
			}

			fmt.Println()
			for _, p := range pets {
				active := "  "
				if p.Active {
					active = "> "
				}
				shiny := ""
				if p.Shiny {
					shiny = " *SHINY*"
				}
				stars := ""
				if p.Prestige > 0 {
					stars = " " + strings.Repeat("*", p.Prestige)
				}
				fmt.Printf("%s%-12s  Lv.%-3d  %-14s  %-9s  Stage %d%s%s\n",
					active, p.Name, p.Level, p.Species, p.Rarity, p.Evolution, stars, shiny)
			}
			fmt.Println()
			return nil
		},
	}
}

func switchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch your active pet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetPetByName(args[0])
			if err != nil {
				fmt.Printf("No pet named '%s' found. Run 'pet list' to see your pets.\n", args[0])
				return nil
			}

			if pet.Active {
				fmt.Printf("%s is already your active pet!\n", pet.Name)
				return nil
			}

			if err := s.SetActivePet(pet.ID); err != nil {
				return err
			}

			fmt.Printf("Switched to %s! (Lv.%d %s)\n", pet.Name, pet.Level, pet.Species)
			return nil
		},
	}
}

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <name>",
		Short: "Permanently delete a pet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			pet, err := s.GetPetByName(args[0])
			if err != nil {
				fmt.Printf("No pet named '%s' found.\n", args[0])
				return nil
			}

			fmt.Printf("\nPermanently delete %s? (Lv.%d %s %s)\n", pet.Name, pet.Level, pet.Rarity, pet.Species)
			fmt.Println("This cannot be undone.")
			fmt.Print("\nType the pet's name to confirm: > ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != pet.Name {
				fmt.Println("Names don't match. Cancelled.")
				return nil
			}

			if err := s.DeletePet(pet.ID); err != nil {
				return err
			}

			fmt.Printf("\n%s is gone forever.\n\n", pet.Name)

			// If we deleted the active pet, activate another one if available
			if pet.Active {
				pets, err := s.GetAllPets()
				if err == nil && len(pets) > 0 {
					s.SetActivePet(pets[0].ID)
					fmt.Printf("%s is now your active pet.\n", pets[0].Name)
				}
			}

			return nil
		},
	}
}

func logCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Show recent commit history with XP and drops",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			defer s.Close()

			commits, err := s.GetRecentCommits(20)
			if err != nil {
				return err
			}

			if len(commits) == 0 {
				fmt.Println("No commits yet. Start committing to earn XP!")
				return nil
			}

			fmt.Println()
			fmt.Printf("  %-12s  %-6s  %-5s  %-5s  %-20s  %s\n",
				"PET", "XP", "LINES", "FILES", "DROP", "TIME")
			fmt.Println("  " + strings.Repeat("-", 70))

			for _, c := range commits {
				dropDisplay := ""
				if c.DropItem != "" && c.DropItem != "nothing" {
					dropDisplay = drop.DisplayName(c.DropItem)
					if dropDisplay == "" {
						dropDisplay = c.DropItem
					}
				}

				timeAgo := timeAgoStr(c.CreatedAt)

				fmt.Printf("  %-12s  +%-5d  %-5d  %-5d  %-20s  %s\n",
					c.PetName, c.XPEarned, c.LinesChanged, c.FilesTouched, dropDisplay, timeAgo)
			}
			fmt.Println()
			return nil
		},
	}
}

func timeAgoStr(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
