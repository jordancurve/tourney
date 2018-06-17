package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const debug = false // Whether to show debug messages.

func main() {
	nSim := 100
	nWin := 0
	nPlayer := 32
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nRound := 40
	for i := 0; i < nSim; i++ {
		strength := make([]float64, nPlayer)
		for i := 0; i < nPlayer; i++ {
			strength[i] = math.RoundToEven(r.NormFloat64()*1000) / 1000.0
		}
		//fmt.Printf("%v\n", strength)
		//if tourneyOk(r, strength, nRound, firstPlayerWinProb) {
		if tourneyOk(r, strength, nRound, strongestPlayerWins) {
			nWin++
		}
	}
	fmt.Printf("win probability = %d/%d (%g)\n",
		nWin, nSim, float64(nWin)/float64(nSim))
}

func tourneyOk(r *rand.Rand, strength []float64, nRound int, fpwb func(s0, s1 float64) float64) bool {
	score := make([]int, len(strength)) // Each player's score.
	seed := getSeed(strength)
	var dist []int
	for i := 0; i < nRound; i++ {
		if debug {
			showScores(strength, score)
		}
		matchups := pairings(score, seed)
		if debug {
			showMatchups(strength, matchups)
		}
		playRound(r, matchups, strength, score, fpwb)
		dist = getDistances(intArrayToFloat64(score), strength)
	}
	if debug {
		showScores(strength, score)
		fmt.Printf("dist=%v\n", dist)
	}
	for _, v := range dist {
		if v > 4 {
			return false
		}
	}
	return true
}

func showScores(strength []float64, score []int) {
	scores := []string{}
	for i := 0; i < len(strength); i++ {
		scores = append(scores, fmt.Sprintf("%2d:%d ", int(strength[i]), int(score[i])))
	}
	sort.Strings(scores)
	for i := 0; i < len(scores)/2; i++ {
		scores[i], scores[len(scores)-i-1] = scores[len(scores)-i-1], scores[i]
	}
	s := strings.Join(scores, " ")
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)
	fmt.Printf("score=%s\n", s)
}

func showMatchups(strength []float64, matchups []int) {
	fmt.Printf("matchups  =")
	for i := 0; i < len(matchups); i += 2 {
		fmt.Printf("%dv%d ", int(strength[matchups[i]]), int(strength[matchups[i+1]]))
	}
	fmt.Printf("\n")
}

// Precondition: elements of x are exactly representable as floats.
// absolute value < 1e9 works.
func intArrayToFloat64(x []int) []float64 {
	y := make([]float64, len(x))
	for i, v := range x {
		y[i] = float64(v)
	}
	return y
}

func strongestPlayerWins(s0, s1 float64) float64 {
	if s0 >= s1 {
		return 1
	}
	return 0
}

// Probability higher-rated player wins.
func firstPlayerWinProb(s0, s1 float64) float64 {
	return clamp(0.5+(s0-s1)*0.4/1.28, 0, 1)
}

func clamp(a, lo, hi float64) float64 {
	if a < lo {
		return lo
	}
	if a > hi {
		return hi
	}
	return a
}

// Given strength of player 0 and player 1, return true iff the
// first player wins
func firstPlayerWins(r *rand.Rand, s0, s1 float64, fpwb func(s0, s1 float64) float64) bool {
	return r.Float64() <= fpwb(s0, s1)
}

// playRound plays all the games in each round according to the opponent
// table.
func playRound(r *rand.Rand, matchups []int, strength []float64, score []int, fpwb func(s0, s1 float64) float64) {
	for i := 0; i < len(matchups); i += 2 {
		// 1 point for a win, 0 points for a loss.
		if firstPlayerWins(r, strength[matchups[i]], strength[matchups[i+1]], fpwb) {
			score[matchups[i]]++
		} else {
			score[matchups[i+1]]++
		}
	}
}

// pairings returns a permutation p of 0..N-1 where N is the number of players such that
// for even i < N, p[i] plays p[i+1].
/*
1. Sort all the players by score from best to worst, breaking ties by seed.
The goal of this sorting is twofold:
- to have all the players with the same score be grouped together in the list
- to have the players within each score group be sorted by seed order (best to worst).

Definition: Player P's "first choice opponent" is the Nth player (0-based) in P's group, where N = (size of group) - (Ps 0-based position in group) - 1. So in a group of 4, for the player numbered 0 (within the group), their first choice opponent is the player numbered 3 in the group.

2. For each player P in the list, starting at the top of the list, their opponent O is the player nearest (with ties broken down) to their first choice opponent for whom all of the following are true:

- O hasn't yet been selected this round
- P has previously played O at most floor(2*numPreviousRounds/numPlayers) times. [not implemented yet]
*/
func pairings(score []int, seed []int) []int {
	players := []int{}
	for i := 0; i < len(score); i++ {
		players = append(players, i)
	}
	sort.Slice(players, func(a, b int) bool {
		a, b = players[a], players[b]
		if score[a] > score[b] {
			return true
		} else if score[a] < score[b] {
			return false
		}
		return seed[a] < seed[b]
	})
	start := 0 // first in group
	group := score[players[0]]
	selected := make([]bool, len(players))
	pairings := []int{}
	for i := 0; i < len(players); i++ {
		if score[players[i]] != group {
			pairings = append(pairings, findOpponentsForGroup(selected, start, i)...)
			group = score[players[i]]
			start = i
		}
	}
	pairings = append(pairings, findOpponentsForGroup(selected, start, len(players))...)
	for i := 0; i < len(pairings); i++ {
		pairings[i] = players[pairings[i]]
	}
	return pairings
}

func findOpponentsForGroup(selected []bool, start, end int) []int {
	pairings := []int{}
	for i := start; i < end; i++ {
		if selected[i] {
			continue
		}
		selected[i] = true
		// SSBM-style
		// firstChoiceOpponent := end-(i-start)-1)
		n := end - start // group size

		// Non-SSBM.
		// in grp of 4, 0 plays 2; 1 plays 3; 2 plays 0, 3 plays 1.
		p := i - start
		firstChoiceOpponent := (p+n/2)%n + start
		//fmt.Printf("start=%d end=%d n=%d i=%d fco=%d\n", start, end, n, i, firstChoiceOpponent)
		o, err := findNearestUnselected(selected, firstChoiceOpponent)
		if err != nil {
			panic(err)
		}
		selected[o] = true
		pairings = append(pairings, i, o)
	}
	return pairings
}

func findNearestUnselected(selected []bool, i int) (int, error) {
	var left, right = i, i
	if !(i >= 0 && i < len(selected)) {
		return -1, fmt.Errorf("starting point (%d) out of range", i)
	}
	for {
		if right < len(selected) {
			if !selected[right] {
				return right, nil
			}
			right++
		} else if left >= 0 {
			if !selected[left] {
				return left, nil
			}
			left--
		} else {
			return -1, fmt.Errorf("unable to find unselected near %d (selected=%v)\n", i, selected)
		}
	}
}

// seed returns an array s[] such that s[i] is the seed of player i, with 0 being the first (best) seed.
func getSeed(strength []float64) []int {
	seed := make([]int, len(strength))
	ps := []int{}
	for i := 0; i < len(strength); i++ {
		ps = append(ps, i)
	}
	sort.Slice(ps, func(a, b int) bool {
		return strength[ps[a]] > strength[ps[b]]
	})
	for i, p := range ps {
		seed[p] = i
	}
	return seed
}

/*
Make an array a numbered from 0 to N. Make one copy of the array sorted by x and the other by y.
Let i be an integer in [0,N). Let pos(i, x) be the position of i in x. Let pos(i, y) be the position
of i in y. Let dist[i] = abs(pos(i, x) - pos(i, y)).
*/
func getDistances(x, y []float64) []int {
	N := len(x)
	a := make([]int, N)
	b := make([]int, N)
	for i := 0; i < N; i++ {
		a[i] = i
		b[i] = i
	}
	sort.Slice(a, func(f, g int) bool { return x[a[f]] > x[a[g]] })
	sort.Slice(b, func(f, g int) bool { return y[b[f]] > y[b[g]] })
	apos := getPosArray(a)
	bpos := getPosArray(b)
	if debug {
		fmt.Printf("sorted by score=%v\n", a)
		fmt.Printf("sorted by strength=%v\n", b)
	}
	dist := make([]int, N)
	for i := 0; i < N; i++ {
		dist[i] = abs(apos[i] - bpos[i])
	}
	return dist
}

/*
Given an array x of N ints that is a permutation of the first N nonnegative integers,
return another array x' that is a permutation of x such that x'[i] = k just when
x[k] = i. That is, x'[v] gives the *position* in array x of the value v.
*/
func getPosArray(x []int) []int {
	y := make([]int, len(x))
	for i, v := range x {
		y[v] = i
	}
	return y
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
