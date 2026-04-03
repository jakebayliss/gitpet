package evolution

type Stage struct {
	Level         int
	StatMultiplier float64
}

var Stages = []Stage{
	{Level: 0, StatMultiplier: 1.0},   // Stage 1
	{Level: 30, StatMultiplier: 1.3},  // Stage 2
	{Level: 60, StatMultiplier: 1.6},  // Stage 3
	{Level: 90, StatMultiplier: 2.0},  // Stage 4
}

func CanEvolve(currentStage, level int, hasEvolutionStone bool) bool {
	nextStage := currentStage + 1
	if nextStage > 4 {
		return false
	}
	if !hasEvolutionStone {
		return false
	}
	return level >= Stages[nextStage-1].Level
}

func NextStageLevel(currentStage int) int {
	nextStage := currentStage + 1
	if nextStage > 4 {
		return 0
	}
	return Stages[nextStage-1].Level
}

func EffectiveStat(baseStat int, stage int) int {
	if stage < 1 || stage > 4 {
		return baseStat
	}
	return int(float64(baseStat) * Stages[stage-1].StatMultiplier)
}
