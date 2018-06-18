package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// A player in the tournament.
type Player struct {
	strength      float64
	score         int // number of wins in the tournament so far
	seed          int // 0 is best seed
	gamesVsPlayer []int
}

type Players []Player

type Tourney struct {
	Rand               *rand.Rand
	NRound             int // number of rounds
	CurRound           int // current round (0-based)
	FirstPlayerWinProb func(s0, s1 float64) float64
	Players            Players
	Matchups           []int
}

const debug = false // Whether to show debug messages.

func main() {
	nSim := 100 // Number of simulations to run.
	nOk := 0    // Number of tournaments that ordered players acceptably.
	nPlayer := 16
	t := &Tourney{
		NRound:             40,
		FirstPlayerWinProb: firstPlayerWinProb,
		Rand:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for i := 0; i < nSim; i++ {
		t.initPlayersWithRandomStrength(nPlayer)
		if t.ok() {
			nOk++
		}
	}
	fmt.Printf("win probability = %d/%d (%g)\n",
		nOk, nSim, float64(nOk)/float64(nSim))
}

func (t *Tourney) initPlayersWithRandomStrength(nPlayer int) {
	t.Players = []Player{}
	for i := 0; i < nPlayer; i++ {
		t.Players = append(t.Players, Player{
			strength: math.RoundToEven(t.Rand.NormFloat64()*1000) / 1000.0,
		})
	}
	t.Players.initGamesVsPlayer()
}

func (ps Players) initGamesVsPlayer() {
	for i := range ps {
		ps[i].gamesVsPlayer = make([]int, len(ps))
	}
}

func (t *Tourney) ok() bool {
	ps := t.Players
	strength := ps.strengths()
	for i, s := range getSeed(strength) {
		ps[i].seed = s
	}
	for i := 0; i < t.NRound; i++ {
		if debug {
			ps.showScores()
		}
		t.Matchups = ps.pairings()
		if debug {
			ps.showMatchups()
		}
		t.playRound()
	}
	distance := ps.getDistances()
	if debug {
		ps.showScores()
		fmt.Printf("distance=%v\n", distance)
	}
	for _, v := range distance {
		if v > 4 {
			return false
		}
	}
	return true
}

func (ps Players) sortByMatchups(matchups []int) {
	nps := make([]Player, len(ps))
	copy(nps, ps)
	for i, v := range matchups {
		ps[i] = nps[v]
	}
}

func (ps Players) strengths() []float64 {
	s := []float64{}
	for _, p := range ps {
		s = append(s, p.strength)
	}
	return s
}

func (ps Players) showScores() {
	scores := []string{}
	for i := 0; i < len(ps); i++ {
		scores = append(scores, fmt.Sprintf("%2s:%d ", ps[i], int(ps[i].score)))
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

func (ps Players) showMatchups() {
	fmt.Printf("matchups  =")
	for i := 0; i < len(ps); i += 2 {
		fmt.Printf("%sv%s ", ps[i], ps[i+1])
	}
	fmt.Printf("\n")
}

func (p Player) String() string {
	return fmt.Sprintf("%d", int(p.strength))
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

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return a
}

// playRound plays all the games in each round according to the opponent
// table.
func (t *Tourney) playRound() {
	for i := 0; i < len(t.Matchups); i += 2 {
		t.playMatch(t.Matchups[i], t.Matchups[i+1])
	}
}

// playMatch plays a single game between players i & j.
func (t *Tourney) playMatch(i, j int) {
	ps := t.Players
	ps[i].gamesVsPlayer[j]++
	ps[j].gamesVsPlayer[i]++
	if t.Rand.Float64() <= t.FirstPlayerWinProb(ps[i].strength, ps[j].strength) {
		ps[i].score++
	} else {
		ps[j].score++
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

Picture an edge weighted complete graph in which nodes are players and each edge is labeled with the number of games played between those two players. To be valid, the graph must not allow any two edges coming out from a node to differ by more than 1. That is, if L is the least of the labels of every edge coming out of N and G is the greatest, then we must have G - L ≤ 1 at all times.
*/
func (ps Players) pairings() []int {
	pis := []int{} // Player-indexes: list of indexes into ps.
	for i := 0; i < len(ps); i++ {
		pis = append(pis, i)
	}
	sort.Slice(pis, func(a, b int) bool { return orderByScoreAndSeed(ps[pis[a]], ps[pis[b]]) })
	start := 0 // first in group
	group := ps[pis[0]].score
	selected := make([]bool, len(pis))
	pairings := []int{}
	for i := 0; i < len(pis); i++ {
		if ps[pis[i]].score != group {
			pairings = append(pairings, findOpponentsForGroup(selected, start, i)...)
			group = ps[pis[i]].score
			start = i
		}
	}
	pairings = append(pairings, findOpponentsForGroup(selected, start, len(pis))...)
	for i := 0; i < len(pairings); i++ {
		pairings[i] = pis[pairings[i]]
	}
	return pairings
}

func orderByScoreAndSeed(a, b Player) bool {
	return a.score > b.score || (a.score == b.score && a.seed < b.seed)
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
func (ps Players) getDistances() []int {
	N := len(ps)
	a := make([]int, N)
	b := make([]int, N)
	for i := 0; i < N; i++ {
		a[i] = i
		b[i] = i
	}
	sort.Slice(a, func(f, g int) bool { return ps[a[f]].score > ps[a[g]].score })
	sort.Slice(b, func(f, g int) bool { return ps[b[f]].strength > ps[b[g]].strength })
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
