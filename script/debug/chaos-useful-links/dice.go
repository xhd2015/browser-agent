package main

import "math/rand"

// Dice picks the next chaos op under tab/seed constraints.
// Weights match the brainstormed Markov table.
type Dice struct {
	rng *rand.Rand
}

func newDice(seed int64) *Dice {
	return &Dice{rng: rand.New(rand.NewSource(seed))}
}

type weightedOp struct {
	op     OpName
	weight float64
}

// pickOp selects an operation given current harness tab count.
// contentTabs excludes the session /go page.
func (d *Dice) pickOp(contentTabs int, maxTabs int) OpName {
	candidates := []weightedOp{
		{opEvalIdentity, 0.20},
		{opScreenshot, 0.15},
		{opSessionInfo, 0.10},
		{opLogs, 0.05},
	}
	if contentTabs < maxTabs {
		candidates = append(candidates, weightedOp{opOpenSeed, 0.20})
	}
	if contentTabs >= 1 {
		candidates = append(candidates, weightedOp{opNavigateSeed, 0.15})
	}
	if contentTabs >= 2 {
		candidates = append(candidates,
			weightedOp{opBackgroundEval, 0.10},
			weightedOp{opRacePair, 0.05},
		)
	}
	// Bootstrap: no content tabs yet → must open.
	if contentTabs == 0 {
		return opOpenSeed
	}

	var total float64
	for _, c := range candidates {
		total += c.weight
	}
	if total <= 0 {
		return opSessionInfo
	}
	r := d.rng.Float64() * total
	var acc float64
	for _, c := range candidates {
		acc += c.weight
		if r <= acc {
			return c.op
		}
	}
	return candidates[len(candidates)-1].op
}

func (d *Dice) pickSeed(seeds []Seed) Seed {
	if len(seeds) == 0 {
		return Seed{}
	}
	return seeds[d.rng.Intn(len(seeds))]
}

func (d *Dice) pickTab(tabs []TabState) TabState {
	if len(tabs) == 0 {
		return TabState{}
	}
	return tabs[d.rng.Intn(len(tabs))]
}

// pickBackgroundTab returns a non-active content tab when possible.
// activeTabID may be 0 if unknown; then any other tab is fine.
func (d *Dice) pickBackgroundTab(tabs []TabState, activeTabID int64) TabState {
	if len(tabs) == 0 {
		return TabState{}
	}
	var others []TabState
	for _, t := range tabs {
		if t.TabID != activeTabID {
			others = append(others, t)
		}
	}
	if len(others) == 0 {
		// Fall back to any tab that is not the first (best-effort).
		if len(tabs) >= 2 {
			return tabs[1]
		}
		return tabs[0]
	}
	return others[d.rng.Intn(len(others))]
}
