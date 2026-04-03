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

			// First pet is always common
			tier := gen.TierCommon
			if hasPet {
				fmt.Println("You already have a pet! Additional pets come from egg drops.")
				return nil
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

			// Auto-install global git hook
			if err := installGlobalHook(); err != nil {
				fmt.Printf("Warning: couldn't install global git hook: %v\n", err)
				fmt.Println("You can try manually with 'pet init --global'")
			} else {
				fmt.Println("Global git hook installed — every commit in any repo feeds your pet!")
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

			fmt.Print(ui.RenderPetCard(pet))
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

			canEvolve := evolution.CanEvolve(pet.Evolution, pet.Level, hasStone)
			nextLevel := evolution.NextStageLevel(pet.Evolution)

			if !canEvolve {
				fmt.Printf("%s cannot evolve yet.\n", pet.Name)
				if pet.Level < nextLevel {
					fmt.Printf("  Level required: %d (current: %d)\n", nextLevel, pet.Level)
				}
				if !hasStone {
					fmt.Println("  Missing: Evolution Stone (keep committing for drops!)")
				}
				return nil
			}

			// Consume the stone
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
