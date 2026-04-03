package drop

import (
	"math/rand"
	"time"
)

type Drop struct {
	Item   string
	Weight int
}

var Table = []Drop{
	{Item: "nothing", Weight: 907},
	{Item: "xp_crystal", Weight: 60},
	{Item: "title_scroll", Weight: 20},
	{Item: "evolution_stone", Weight: 8},
	{Item: "rare_egg", Weight: 4},
	{Item: "legendary_egg", Weight: 1},
}

var Titles = []string{
	"the Relentless",
	"Bug Slayer",
	"the Unbreakable",
	"Code Devourer",
	"Pixel Pusher",
	"the Caffeinated",
	"Merge Conflict Survivor",
	"the Nocturnal",
	"Stack Overflow",
	"Segfault Whisperer",
	"the Refactored",
	"Commit Goblin",
	"the Untested",
	"Dependency Hell Escapee",
	"the Deprecated",
	"404 Not Found",
	"the Verbose",
	"Null Pointer",
	"the Async",
	"Memory Leak",
}

func Roll() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	totalWeight := 0
	for _, d := range Table {
		totalWeight += d.Weight
	}

	roll := r.Intn(totalWeight)
	cumulative := 0
	for _, d := range Table {
		cumulative += d.Weight
		if roll < cumulative {
			return d.Item
		}
	}
	return "nothing"
}

func RandomTitle() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Titles[r.Intn(len(Titles))]
}

func DisplayName(item string) string {
	switch item {
	case "xp_crystal":
		return "XP Crystal (+200 XP)"
	case "title_scroll":
		return "Title Scroll"
	case "evolution_stone":
		return "Evolution Stone"
	case "rare_egg":
		return "*** RARE EGG ***"
	case "legendary_egg":
		return "***** LEGENDARY EGG *****"
	case "nothing":
		return ""
	default:
		return item
	}
}
