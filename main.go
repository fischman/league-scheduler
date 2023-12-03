// Local Variables:
// compile-command: "LD_LIBRARY_PATH=$HOME/go/src/github.com/draffensperger/golp/lpsolve CGO_CFLAGS=\"-I$HOME/go/src/github.com/draffensperger/golp/lpsolve\" CGO_LDFLAGS=\"-L$HOME/go/src/github.com/draffensperger/golp/lpsolve -llpsolve55 -lm -ldl -lcolamd\" go run main.go"
// End:

package main

import (
	"fmt"
	"log"
	"math"

	"github.com/draffensperger/golp"
)

/*
Variables:
   P(m,d,t) = 1 if team t plays on match m on day d, 0 otherwise.
   - there are three matches per day: 0 <= m <= 2
   - there are six days per season: 0 <= d <= 5
   - there are six teams: 0 <= t <= 5
   M(m,d,t1,t2) = 1 if team t1 and t2 are playing in match m on day d, 0 otherwise.
Constraints:
   - each match has exactly two teams playing
   - each team plays exactly one match per day
   - each team plays every other team at least once
   - each team plays twice in each time slot
   - M & P relate as expected
Maximize:
   - dummy function, only care if we can find a feasible solution to the above.
*/

const (
	numMatches = 3
	numDays    = 6
	numTeams   = 6
	numPVars   = numMatches * numDays * numTeams
	numMVars   = numMatches * numDays * numTeams * numTeams
	numVars    = numPVars + numMVars
)

func pVarsFromIndex(i int) (match, day, team int) {
	team = i % numTeams
	i /= numTeams
	day = i % numDays
	i /= numDays
	match = i
	return
}

func indexFromPVars(m, d, t int) int {
	return t + d*numDays + m*numDays*numTeams
}

func mVarsFromIndex(i int) (match, day, t1, t2 int) {
	i -= numPVars
	t2 = i % numTeams
	i /= numTeams
	match, day, t1 = pVarsFromIndex(i)
	return
}

func indexFromMVars(m, d, t1, t2 int) int {
	return numPVars + m*numDays*numTeams*numTeams + d*numTeams*numTeams + t1*numTeams + t2
}

func league() {
	lp := golp.NewLP(0, numVars)
	for i := 0; i < numPVars; i++ {
		m, d, t := pVarsFromIndex(i)
		lp.SetColName(i, fmt.Sprintf("match %v, day %v, team %v", m, d, t))
		lp.SetBinary(i, true)
	}
	for m := 0; m < numMatches; m++ {
		for d := 0; d < numDays; d++ {
			for t1 := 0; t1 < numTeams; t1++ {
				for t2 := 0; t2 < numTeams; t2++ {
					i := indexFromMVars(m, d, t1, t2)
					lp.SetColName(i, fmt.Sprintf("M(%v,%v,%v,%v)", m, d, t1, t2))
					lp.SetBinary(i, true)
				}
			}
		}
	}

	// Constraint: M vars are symmetric.
	for t1 := 0; t1 < numTeams-1; t1++ {
		for t2 := t1 + 1; t2 < numTeams; t2++ {
			for m := 0; m < numMatches; m++ {
				for d := 0; d < numDays; d++ {
					lp.AddConstraintSparse([]golp.Entry{
						golp.Entry{Col: indexFromMVars(m, d, t1, t2), Val: 1},
						golp.Entry{Col: indexFromMVars(m, d, t2, t1), Val: -1},
					}, golp.EQ, 0)
				}
			}
		}
	}

	// Constraint: relate the P & M vars.
	for t1 := 0; t1 < numTeams; t1++ {
		for t2 := 0; t2 < numTeams; t2++ {
			for m := 0; m < numMatches; m++ {
				for d := 0; d < numDays; d++ {
					// This trick for modeling an AND relationship is from
					// https://math.stackexchange.com/q/2893783: to model v1*v2
					// (a.k.a. v1 AND v2), introduce a variable m and constrain
					// 3 ways:
					// m <= v1
					lp.AddConstraintSparse([]golp.Entry{
						golp.Entry{Col: indexFromMVars(m, d, t1, t2), Val: 1},
						golp.Entry{Col: indexFromPVars(m, d, t1), Val: -1},
					}, golp.LE, 0)
					// m <= v2
					lp.AddConstraintSparse([]golp.Entry{
						golp.Entry{Col: indexFromMVars(m, d, t1, t2), Val: 1},
						golp.Entry{Col: indexFromPVars(m, d, t2), Val: -1},
					}, golp.LE, 0)
					// m >= v1+v2-1
					lp.AddConstraintSparse([]golp.Entry{
						golp.Entry{Col: indexFromMVars(m, d, t1, t2), Val: 1},
						golp.Entry{Col: indexFromPVars(m, d, t1), Val: -1},
						golp.Entry{Col: indexFromPVars(m, d, t2), Val: -1},
					}, golp.GE, -1)
				}
			}
		}
	}

	// Constraint: each team plays every other team at least once and at most twice, and never itself.
	for t1 := 0; t1 < numTeams; t1++ {
		for t2 := 0; t2 < numTeams; t2++ {
			row := []golp.Entry{}
			for m := 0; m < numMatches; m++ {
				for d := 0; d < numDays; d++ {
					row = append(row, golp.Entry{Col: indexFromMVars(m, d, t1, t2), Val: 1})
				}
			}
			if t1 == t2 {
				lp.AddConstraintSparse(row, golp.EQ, 0)
			} else {
				lp.AddConstraintSparse(row, golp.LE, 2)
				lp.AddConstraintSparse(row, golp.GE, 1)
			}
		}
	}

	// Constraint: each match has exactly two teams playing.
	for m := 0; m < numMatches; m++ {
		for d := 0; d < numDays; d++ {
			row := []golp.Entry{}
			for t := 0; t < numTeams; t++ {
				row = append(row, golp.Entry{Col: indexFromPVars(m, t, d), Val: 1})
			}
			lp.AddConstraintSparse(row, golp.EQ, 2)
		}
	}

	// Constraint: each team plays exactly one match per day.
	for t := 0; t < numTeams; t++ {
		for d := 0; d < numDays; d++ {
			row := []golp.Entry{}
			for m := 0; m < numMatches; m++ {
				row = append(row, golp.Entry{Col: indexFromPVars(m, t, d), Val: 1})
			}
			lp.AddConstraintSparse(row, golp.EQ, 1)
		}
	}

	// Constraint: each team plays twice in each timeslot across all days.
	for t := 0; t < numTeams; t++ {
		for m := 0; m < numMatches; m++ {
			row := []golp.Entry{}
			for d := 0; d < numDays; d++ {
				row = append(row, golp.Entry{indexFromPVars(m, t, d), 1})
			}
			lp.AddConstraintSparse(row, golp.EQ, 2)
		}
	}

	// Objective: dummy to find any feasible solution to above.
	obj := make([]float64, numVars)
	obj[0] = 1
	lp.SetObjFn(obj)
	if got := lp.Solve(); got == golp.INFEASIBLE {
		log.Fatalf("INFEASIBLE")
	}

	schedule := lp.Variables()
	for m := 0; m < numMatches; m++ {
		for d := 0; d < numDays; d++ {
			for t := 0; t < numTeams; t++ {
				i := indexFromPVars(m, d, t)
				if schedule[i] > 0 {
					fmt.Printf("%v", string('A'+byte(t)))
				}
			}
			fmt.Printf(" ")
		}
		fmt.Println("")
	}
}

// This is from the golp docs and serves to confirm needed dependencies are found and working.
func sample() {
	lp := golp.NewLP(0, 2)
	lp.AddConstraint([]float64{110.0, 30.0}, golp.LE, 4000.0)
	lp.AddConstraint([]float64{1.0, 1.0}, golp.LE, 75.0)
	lp.SetObjFn([]float64{143.0, 60.0})
	lp.SetMaximize()

	if got, want := lp.Solve(), golp.SolutionType(golp.OPTIMAL); got != want {
		log.Fatalf("lp.Solve: got: %v, want: %v", got, want)
	}
	floatSliceAlmostEqual := func(a, b []float64) bool {
		const eps = 1e-4
		if len(a) != len(b) {
			return false
		}
		for i, av := range a {
			if math.Abs(b[i]-av) > eps {
				return false
			}
		}
		return true
	}
	if got, want := lp.Variables(), []float64{21.875, 53.125}; !floatSliceAlmostEqual(got, want) {
		log.Fatalf("lp.Variables: got: %v want: %v", got, want)
	}
	if got, want := lp.Objective(), 6315.625; got != want {
		log.Fatalf("lp.Objective: got: %v, want: %v", got, want)
	}
}

func main() {
	// Test round-tripping through both kinds of variables and indexing helpers.
	for i := 0; i < numVars; i++ {
		if i < numPVars {
			m, d, t := pVarsFromIndex(i)
			if want, got := i, indexFromPVars(m, d, t); got != want {
				log.Fatalf("Failed to round-trip P vars; want: %v got: %v", want, got)
			}
		} else {
			m, d, t1, t2 := mVarsFromIndex(i)
			if want, got := i, indexFromMVars(m, d, t1, t2); got != want {
				log.Fatalf("Failed to round-trip M vars; want: %v got: %v", want, got)
			}
		}
	}
	for i := 0; i < 108; i++ {
		m, d, t := pVarsFromIndex(i)
		if want, got := i, indexFromPVars(m, d, t); got != want {
			log.Fatalf("Failed to round-trip; want: %v got: %v", want, got)
		}
	}

	sample()
	league()
}
