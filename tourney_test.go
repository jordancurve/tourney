package main

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func TestSeed(t *testing.T) {
	f := func(strength []float64) bool {
		s := getSeed(strength)
		if len(s) != len(strength) {
			return false
		}
		p := make([]int, len(s))
		for i := 0; i < len(s); i++ {
			p[s[i]] = i
		}
		for i := 0; i < len(p)-1; i++ {
			if strength[p[i]] < strength[p[i+1]] {
				fmt.Printf("i=%d p[i]=%d p[i+1]=%d strength[%d]=%g strength[%d]=%g\n",
					i, p[i], p[i+1], p[i], strength[p[i]], p[i+1], strength[p[i+1]])
				return false
			}
		}
		seen := map[int]bool{}
		for i := 0; i < len(s); i++ {
			if seen[i] {
				return false
			}
			seen[i] = true
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestFindNearestUnselected(t *testing.T) {
	cases := []struct {
		selected []bool
		start    int
		want     int
	}{
		{[]bool{false}, 0, 0},
		{[]bool{false, true}, 0, 0},
		{[]bool{false, true}, 1, 0},
		{[]bool{false, true, false}, 1, 2},
		{[]bool{false, true, true}, 1, 0},
		{[]bool{false, true, true}, 2, 0},
		{[]bool{false, true, true}, 0, 0},
		{[]bool{true, false, true, true, true, true, true, false}, 4, 7},
		{[]bool{true, false, true, true, true, true, true, true}, 4, 1},
	}
	for _, c := range cases {
		got, err := findNearestUnselected(c.selected, c.start)
		if err != nil {
			t.Errorf("findNearestUnselected(%v, %d) returned unexpected error %v", c.selected, c.start, err)
		}
		if got != c.want {
			t.Errorf("findNearestUnselected(%v, %d)=%d; want %d", c.selected, c.start, got, c.want)
		}
	}
}

func TestFindOpponentsForGroup(t *testing.T) {
	cases := []struct {
		selected   []bool
		start, end int
		want       []int
	}{
		{[]bool{false, false}, 0, 2, []int{0, 1}},
		//SSBM:
		//{[]bool{false, false, false, false}, 0, 4, []int{0, 3, 1, 2}},
		//{[]bool{false, true, false, false, false}, 1, 5, []int{2, 3, 4, 0}},
		{[]bool{false, false, false, false}, 0, 4, []int{0, 2, 1, 3}},
		{[]bool{false, true, false, false, false}, 1, 5, []int{2, 4, 3, 0}},
	}
	for _, c := range cases {
		got := findOpponentsForGroup(c.selected, c.start, c.end)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("findOpponentsForGroup(%v, %v, %v)=%v; want %v", c.selected, c.start, c.end, got, c.want)
		}
	}
}

func TestPairings(t *testing.T) {
	cases := []struct {
		score []int
		seed  []int
		want  []int
	}{
		{[]int{1, 0}, []int{0, 0}, []int{0, 1}},
		{[]int{1, 1}, []int{0, 1}, []int{0, 1}},
		{[]int{2, 2, 1, 1}, []int{0, 1, 2, 3}, []int{0, 1, 2, 3}},
		{[]int{3, 2, 1, 1}, []int{0, 1, 2, 3}, []int{0, 1, 2, 3}},
		// SSBM-style
		// {[]int{1, 1, 1, 1}, []int{0, 1, 2, 3}, []int{0, 3, 1, 2}},
		// {[]int{1, 1, 1, 1, 0, 0, 0, 0}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 3, 1, 2, 4, 7, 5, 6}},
		{[]int{1, 1, 1, 1}, []int{0, 1, 2, 3}, []int{0, 2, 1, 3}},
		{[]int{1, 1, 1, 1, 0, 0, 0, 0}, []int{0, 1, 2, 3, 4, 5, 6, 7}, []int{0, 2, 1, 3, 4, 6, 5, 7}},
	}
	for _, c := range cases {
		p := make([]Player, len(c.score))
		for i := 0; i < len(p); i++ {
			p[i].score = c.score[i]
			p[i].seed = c.seed[i]
		}
		got := Players(p).pairings()
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("pairings(%v, %v)=%v; want %v", c.score, c.seed, got, c.want)
		}
	}
}

func TestGetPosArray(t *testing.T) {
	cases := []struct {
		in   []int
		want []int
	}{
		{[]int{}, []int{}},
		{[]int{0}, []int{0}},
		{[]int{0, 1}, []int{0, 1}},
		{[]int{1, 0}, []int{1, 0}},
		{[]int{1, 2, 0}, []int{2, 0, 1}},
	}
	for _, c := range cases {
		got := getPosArray(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getPosArray(%v)=%v; want %v", c.in, got, c.want)
		}
	}
}

func TestAbs(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{-1, 1},
		{0, 0},
		{1, 1},
	}
	for _, c := range cases {
		got := abs(c.in)
		if got != c.want {
			t.Errorf("abs(%v)=%v; want %v", c.in, got, c.want)
		}
	}
}

func TestGetDistances(t *testing.T) {
	cases := []struct {
		score    []int
		strength []float64
		want     []int
	}{
		{[]int{}, []float64{}, []int{}},
		{[]int{0}, []float64{0}, []int{0}},
		{[]int{0, 1}, []float64{0, 1}, []int{0, 0}},
		{[]int{1, 0}, []float64{1, 0}, []int{0, 0}},
		{[]int{0, 1}, []float64{1, 0}, []int{1, 1}},
		{[]int{0, 1, 2}, []float64{2, 1, 0}, []int{2, 0, 2}},
	}
	for _, c := range cases {
		p := make([]Player, len(c.score))
		for i := 0; i < len(p); i++ {
			p[i].score = c.score[i]
			p[i].strength = c.strength[i]
		}
		got := Players(p).getDistances()
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getDistances(%v, %v)=%v; want %v", c.score, c.strength, got, c.want)
		}
	}
}

func TestIntArrayToFloat64(t *testing.T) {
	f := func(x []int) bool {
		for i := range x {
			if x[i] > 1e9 {
				x[i] = 1e9
			} else if x[i] < -1e9 {
				x[i] = -1e9
			}
		}
		y := intArrayToFloat64(x)
		if len(y) != len(x) {
			return false
		}
		for i := 0; i < len(y); i++ {
			if int(y[i]) != x[i] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestFirstPlayerWinProb(t *testing.T) {
	cases := []struct {
		s0, s1, want float64
	}{
		{0, 0, 0.5},
		{1.28, 0, 0.9},
		{-1.28, 0, 0.1},
		{1.6, 0, 1.0},
		{-1.6, -1.3, .40625},
		{1.6, 1.3, .59375},
		{1.5, 1.2, .59375},
	}
	for _, c := range cases {
		got := firstPlayerWinProb(c.s0, c.s1)
		if math.Abs(got-c.want) > 0.00001 {
			t.Errorf("firstPlayerWinProb(%v, %v)=%v; want %v", c.s0, c.s1, got, c.want)
		}
	}
}

func r() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

func TestTourneyOk(t *testing.T) {
	cases := [][]float64{
		{0, 1},
		{0, .5, 1, 1.5},
		{-1, -0.75, -0.5, -0.25, 0, .5, 1, 1.5},
		{9, 7, 5, 3, 8, 6, 1, 0, 2, 4},
		{5, 1, 8, 11, 10, 2, 4, 6, 7, 9, 12, 3},
		{8, 5, 14, 13, 11, 3, 15, 7, 6, 10, 9, 1, 0, 2, 12, 4},
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	}
	for _, c := range cases {
		p := make([]Player, len(c))
		for i, s := range c {
			p[i].strength = s
		}
		ty := Tourney{
			NRound:             1000,
			FirstPlayerWinProb: strongestPlayerWins,
			Players:            Players(p),
			Rand:               r(),
		}
		ty.Players.initGamesVsPlayer()
		if !ty.ok() {
			t.Errorf("(%#v).ok() failed", ty)
		}
	}
}
