package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

	root.AddCommand(hatchCmd())
	root.AddCommand(showCmd())
	root.AddCommand(initCmd())
	root.AddCommand(commitCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func hatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hatch",
		Short: "Hatch a new pet",
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
			fmt.Println("\nRun 'pet show' to see your pet.")
			fmt.Println("Run 'pet init' in a git repo to start earning XP!")
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
				fmt.Println("No pet found! Run 'pet hatch' to get started.")
				return nil
			}

			fmt.Print(ui.RenderPetCard(pet))
			return nil
		},
	}
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Install git hook in current repo (or globally with --global)",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")

			hookScript := "#!/bin/sh\npet commit\n"

			if global {
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
				// Set git global hooks path
				if err := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir).Run(); err != nil {
					return fmt.Errorf("setting global hooks path: %w", err)
				}
				fmt.Printf("Global post-commit hook installed at %s\n", hookPath)
				fmt.Println("All repos will now feed your pet on commit!")
				return nil
			}

			// Local repo hook
			gitDir, err := exec.Command("git", "rev-parse", "--git-dir").Output()
			if err != nil {
				return fmt.Errorf("not a git repository")
			}
			hooksDir := filepath.Join(strings.TrimSpace(string(gitDir)), "hooks")
			if err := os.MkdirAll(hooksDir, 0755); err != nil {
				return err
			}
			hookPath := filepath.Join(hooksDir, "post-commit")

			// Check if hook already exists
			if _, err := os.Stat(hookPath); err == nil {
				existing, _ := os.ReadFile(hookPath)
				if strings.Contains(string(existing), "pet commit") {
					fmt.Println("Hook already installed!")
					return nil
				}
				// Append to existing hook
				f, err := os.OpenFile(hookPath, os.O_APPEND|os.O_WRONLY, 0755)
				if err != nil {
					return err
				}
				defer f.Close()
				f.WriteString("\npet commit\n")
				fmt.Println("Added pet to existing post-commit hook.")
				return nil
			}

			if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
				return err
			}
			fmt.Printf("Post-commit hook installed! Your pet will gain XP on every commit.\n")
			return nil
		},
	}
	cmd.Flags().Bool("global", false, "Install hook globally for all repos")
	return cmd
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

			oldLevel := pet.Level
			pet.XP += result.TotalXP
			newLevel := xp.LevelFromXP(pet.XP)
			leveledUp := newLevel > oldLevel
			pet.Level = newLevel

			if err := s.UpdatePetXP(pet.ID, pet.XP, pet.Level); err != nil {
				return err
			}

			// Record commit
			commit := &store.Commit{
				PetID:        pet.ID,
				Repo:         stats.Repo,
				LinesChanged: stats.LinesChanged,
				FilesTouched: stats.FilesTouched,
				XPEarned:     result.TotalXP,
				DropItem:     "nothing",
				CreatedAt:    time.Now(),
			}
			if err := s.RecordCommit(commit); err != nil {
				return err
			}

			branchBonus := result.BranchBonus > 1.0
			fmt.Print(ui.RenderCommitResult(
				pet.Name, result.TotalXP, pet.XP, pet.Level,
				leveledUp, oldLevel, streakDays, branchBonus, "nothing",
			))

			return nil
		},
	}
}
// test
