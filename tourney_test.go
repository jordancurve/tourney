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
		{[]bool{false, false, false, false}, 0, 4, []int{0, 3, 1, 2}},
		{[]bool{false, true, false, false, false}, 1, 5, []int{2, 3, 4, 0}},
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
		{[]int{1, 1, 1, 1}, []int{0, 1, 2, 3}, []int{0, 3, 1, 2}},
		{[]int{1, 1}, []int{0, 1}, []int{0, 1}},
		{[]int{2, 2, 1, 1}, []int{0, 1, 2, 3}, []int{0, 1, 2, 3}},
		{[]int{3, 2, 1, 1}, []int{0, 1, 2, 3}, []int{0, 1, 2, 3}},
	}
	for _, c := range cases {
		got := pairings(c.score, c.seed)
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
		x, y []float64
		want []int
	}{
		{[]float64{}, []float64{}, []int{}},
		{[]float64{0}, []float64{0}, []int{0}},
		{[]float64{0, 1}, []float64{0, 1}, []int{0, 0}},
		{[]float64{1, 0}, []float64{1, 0}, []int{0, 0}},
		{[]float64{0, 1}, []float64{1, 0}, []int{1, 1}},
		{[]float64{0, 1, 2}, []float64{2, 1, 0}, []int{2, 0, 2}},
	}
	for _, c := range cases {
		got := getDistances(c.x, c.y)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getDistances(%v, %v)=%v; want %v", c.x, c.y, got, c.want)
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

func strongestPlayerWins(s0, s1 float64) float64 {
	if s0 >= s1 {
		return 1
	}
	return 0
}

func TestTourneyOk(t *testing.T) {
	cases := [][]float64{
		{0, 1},
		{0, .5, 1, 1.5},
		{-1, -0.75, -0.5, -0.25, 0, .5, 1, 1.5},
		{-0.178, -0.259, -0.33, -1.076, 0.033, -1.707, 0.466, 1.107, 0.579, -0.825, -0.255, 0.086, 0.309, 0.512, 1.713, -0.777},
		{0.718, -1.027, -0.172, -0.914, -0.077, 1.992, 0.742, -0.242, 0.291, -0.564, 1.089, 0.715, -1.331, -1.085, -1.071, -0.286},
		{0.16, 0.5, 1, 0, -0.4, 1.4, -0.7, -0.6, -0.15, 1.3, 0.8, -1.5},
	}
	for _, d := range cases {
		r := rand.New(rand.NewSource(0))
		if !tourneyOk(r, d, 100, strongestPlayerWins) {
			t.Fatalf("tourneyOk(%v) failed", d)
		}
	}
}
