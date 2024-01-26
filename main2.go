// -*- mode: Go; compile-command: "go run main2.go" -*-

package main

import (
	"fmt"
	"log"
)

type match [2]int

func (m *match) String() string {
	a, b := byte((*m)[0]), byte((*m)[1])
	a, b = min(a, b), max(a, b)
	return string([]byte{'A' + a, 'A' + b})
}

func (m match) overlap(m2 match) bool {
	return m[0] == m2[0] || m[1] == m2[0] || m[0] == m2[1] || m[1] == m2[1]
}

type day [3]match

type slotCounts [3]int

func (s *slotCounts) inc(pos int) {
	(*s)[pos]++
}
func (s *slotCounts) dec(pos int) {
	(*s)[pos]--
}

func (s *slotCounts) String() string {
	return fmt.Sprintf("%v/%v/%v", (*s)[0], (*s)[1], (*s)[2])
}

type freqs map[int]*slotCounts // Map team to number of times they've played in each slot.

func (f *freqs) check(m match, pos int) bool {
	if (*f)[m[0]][pos] == 2 || (*f)[m[1]][pos] == 2 {
		return false
	}
	(*f)[m[0]].inc(pos)
	(*f)[m[1]].inc(pos)
	return true
}

func (f *freqs) remove(m match, pos int) {
	(*f)[m[0]].dec(pos)
	(*f)[m[1]].dec(pos)
}

func addDay(previousDays []day, previousFreqs *freqs, remainingMatches []match) []day {
	if len(remainingMatches) == 0 {
		return previousDays
	}
	for i1, m1 := range remainingMatches {
		if !previousFreqs.check(m1, 0) {
			continue
		}

		for i2, m2 := range remainingMatches {
			if m1.overlap(m2) {
				continue
			}
			if !previousFreqs.check(m2, 1) {
				continue
			}
			for i3, m3 := range remainingMatches {
				if m1.overlap(m3) || m2.overlap(m3) {
					continue
				}
				if !previousFreqs.check(m3, 2) {
					continue
				}

				pd := append(previousDays, day{m1, m2, m3})
				rm := []match{}
				for i := 0; i < len(remainingMatches); i++ {
					if i != i1 && i != i2 && i != i3 {
						rm = append(rm, remainingMatches[i])
					}
				}
				if got := addDay(pd, previousFreqs, rm); got != nil {
					return got
				}

				previousFreqs.remove(m3, 2)
			}
			previousFreqs.remove(m2, 1)
		}
		previousFreqs.remove(m1, 0)
	}
	return nil
}

func main() {
	const numTeams = 6
	matches := []match{}
	freqs := &freqs{}
	for i := 0; i < numTeams-1; i++ {
		for j := i + 1; j < numTeams; j++ {
			matches = append(matches, match{i, j})
		}
	}
	for i := 0; i < numTeams; i++ {
		(*freqs)[i] = &slotCounts{}
	}

	// This generates the first 5 days' worth of matches, with all
	// team-pairs playing once.
	schedule := addDay([]day{}, freqs, matches)
	if got, want := len(schedule), 5; got != want {
		log.Fatalf("len(schedule): got: %v, want %v", got, want)
	}

	// Populate the 6th day by figuring out, for each slot, which teams
	// haven't played twice in that slot yet, and match them up.
	extraMatchSlotToTeams := map[int][]int{}
	for t, s := range *freqs {
		for i := 0; i < len(s); i++ {
			if s[i] < 2 {
				extraMatchSlotToTeams[i] = append(extraMatchSlotToTeams[i], t)
				break
			}
		}
	}
	schedule = append(schedule, day{
		match{extraMatchSlotToTeams[0][0], extraMatchSlotToTeams[0][1]},
		match{extraMatchSlotToTeams[1][0], extraMatchSlotToTeams[1][1]},
		match{extraMatchSlotToTeams[2][0], extraMatchSlotToTeams[2][1]},
	})

	for s := 0; s < 3; s++ {
		for d := 0; d < len(schedule); d++ {
			if d != 0 {
				fmt.Printf(" ")
			}
			fmt.Printf("%v", schedule[d][s].String())
		}
		fmt.Println()
	}

}
