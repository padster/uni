package main

import (
	"fmt"
)

const (
	ROWS = 3
	COLS = 4
)

var D = []int{1, 0, -1, 0, 1}

const (
	DOWN  = 0 // [ 1,  0]
	LEFT  = 1 // [ 0, -1]
	UP    = 2 // [-1,  0]
	RIGHT = 3 // [ 0,  1]
)

type BeliefState [ROWS][COLS]float64

var WALLS = [ROWS][COLS]int{
	{2, 2, 1, -1},
	{2, 4, 1, -1},
	{2, 2, 1, 2},
}

func main() {
	runPOMDP([]int{UP, UP, UP}, []int{2, 2, 2}, uniformBelief())
	runPOMDP([]int{UP, UP, UP}, []int{1, 1, 1}, uniformBelief())
	runPOMDP([]int{RIGHT, RIGHT, UP}, []int{1, 1, -1}, knownState(2, 3))
	runPOMDP([]int{UP, RIGHT, RIGHT, RIGHT}, []int{2, 2, 1, 1}, knownState(1, 1))
}

func runPOMDP(actions []int, observations []int, b *BeliefState) {
	if len(actions) != len(observations) {
		panic("# Actions != # Observations")
	}
	b.Print(0)
	for i := range actions {
		b = updateBelief(b, actions[i], observations[i])
		b.Print(i + 1)
	}
	fmt.Printf("\n\n\n")
}

func updateBelief(b *BeliefState, action int, observation int) *BeliefState {
	result := &BeliefState{}
	for r, row := range result {
		for c := range row {
			result[r][c] = stateProbability(r, c, b, action, observation)
		}
	}
	result.normalize()
	return result
}

// Probability of reaching (rAt, cAt) given a belief state, action, and observation.
func stateProbability(rAt, cAt int, b *BeliefState, action, observation int) float64 {
	if WALLS[rAt][cAt] == 4 { // Unreachable
		return 0.0
	}
	sum := 0.0
	for rFrom, row := range b {
		for cFrom := range row {
			sum += b[rFrom][cFrom] * transitionProability(rAt, cAt, rFrom, cFrom, action)
		}
	}
	return sum * observationProbability(rAt, cAt, observation)
}

// Probability of moving from (rFrom, cFrom) to (rAt, cAt) given an action.
func transitionProability(rAt, cAt int, rFrom, cFrom int, action int) float64 {
	if WALLS[rFrom][cFrom] == -1 || WALLS[rFrom][cFrom] == 4 {
		return 0.0
	}

	sum := 0.0
	if movesTo(rFrom, cFrom, action, rAt, cAt) {
		sum += 0.8
	}
	if movesTo(rFrom, cFrom, (action+1)%4, rAt, cAt) {
		sum += 0.1
	}
	if movesTo(rFrom, cFrom, (action+3)%4, rAt, cAt) {
		sum += 0.1
	}
	return sum
}

// Returns whether moving a given direction from once cell ends up in another.
func movesTo(rFrom, cFrom int, action int, rTo, cTo int) bool {
	return snap(rFrom+D[action], 0, ROWS) == rTo &&
		snap(cFrom+D[action+1], 0, COLS) == cTo
}

// Snaps 'val' into the inclusive range [lo, hi].
func snap(val, lo, hi int) int {
	if val < lo {
		return lo
	} else if val > hi {
		return hi
	} else {
		return val
	}
}

// Probability of observing 1/2/end in state (rAt, cAt).
func observationProbability(rAt, cAt int, observation int) float64 {
	if WALLS[rAt][cAt] == 4 {
		panic("Shouldn't make observations in an unreachable cell...")
	}

	if WALLS[rAt][cAt] == -1 {
		if observation == -1 { // Terminal state, so should always observe 'end' == -1.
			return 1.0
		} else {
			return 0.0
		}
	}

	if WALLS[rAt][cAt] == observation { // Otherwise, depends on how many walls there are.
		return 0.9
	} else if observation != -1 {
		return 0.1
	} else {
		return 0.0 // Can't observe non-end at end.
	}
}

// All non-terminal, non-void states have equal probability.
func uniformBelief() *BeliefState {
	b := &BeliefState{}
	for r, row := range b {
		for c := range row {
			if WALLS[r][c] != -1 && WALLS[r][c] != 4 {
				b[r][c] = 1.0
			}
		}
	}
	b.normalize()
	return b
}

// A given cell starts with 100% probability.
func knownState(xAt, yAt int) *BeliefState {
	// Fix coordinates, (1, 1) = bottom left = [ROWS-1][0]
	rAt, cAt := ROWS-yAt, xAt-1
	b := &BeliefState{}
	b[rAt][cAt] = 1.0
	return b
}

// Normalize a belief state so probabilities sum to 1.
func (b *BeliefState) normalize() {
	sum := 0.0
	for r, row := range b {
		for c := range row {
			if WALLS[r][c] != 4 {
				sum += b[r][c]
			}
		}
	}
	for r, row := range b {
		for c := range row {
			b[r][c] /= sum
		}
	}
}

// Print the belief state for debugging.
func (b *BeliefState) Print(iteration int) {
	fmt.Printf("Belief state %d: \n", iteration)
	for r, row := range b {
		for c := range row {
			fmt.Printf("%.4f\t", b[r][c])
		}
		fmt.Printf("\n")
	}
}


// // Estimates a value using the cumulative rolling average.
// type Estimate struct {
//     count int
//     Value float64
// }
// func (e *Estimate) Update(observed float64) {
//     e.count++
//     e.Value += (observed - e.Value) * (1.0 / float64(e.count))
// }
