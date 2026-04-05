package gen

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jakebayliss/gitpet/internal/store"
)

type Tier string

const (
	TierCommon    Tier = "Common"
	TierRare      Tier = "Rare"
	TierLegendary Tier = "Legendary"
)

var commonSpecies = []string{
	"Axolotl",
	"Capybara",
	"Goblin",
	"Pangolin",
	"Tardigrade",
	"Blobfish",
	"Mantis Shrimp",
	"Nudibranch",
}

var rareSpecies = []string{
	"Salamander",
	"Dire Wolf",
	"Thunderbird",
	"Crystal Fox",
	"Chimera",
	"Shadow Lynx",
}

var legendarySpecies = []string{
	"Phoenix",
	"Leviathan",
	"Basilisk",
	"Void Cat",
	"Kraken",
	"Celestial Beetle",
	"Mimic",
	"Wyrm",
}

var statFloors = map[Tier]int{
	TierCommon:    20,
	TierRare:      50,
	TierLegendary: 80,
}

func Roll(tier Tier) *store.Pet {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var species string
	switch tier {
	case TierRare:
		species = rareSpecies[r.Intn(len(rareSpecies))]
	case TierLegendary:
		species = legendarySpecies[r.Intn(len(legendarySpecies))]
	default:
		species = commonSpecies[r.Intn(len(commonSpecies))]
	}

	floor := statFloors[tier]
	statRange := 20
	stats := store.Stats{
		Wit:     floor + r.Intn(statRange),
		Depth:   floor + r.Intn(statRange),
		Stamina: floor + r.Intn(statRange),
		Luck:    floor + r.Intn(statRange),
		Attune:  floor + r.Intn(statRange),
	}

	shiny := r.Intn(100) == 0 // 1% chance

	now := time.Now()
	return &store.Pet{
		ID:        uuid.New().String(),
		Species:   species,
		Rarity:    string(tier),
		Shiny:     shiny,
		Level:     1,
		XP:        0,
		Evolution: 1,
		Prestige:  0,
		Stats:     stats,
		Active:    true,
		HatchedAt: now,
		UpdatedAt: now,
	}
}

func SpeciesPool(tier Tier) []string {
	switch tier {
	case TierRare:
		return rareSpecies
	case TierLegendary:
		return legendarySpecies
	default:
		return commonSpecies
	}
}
