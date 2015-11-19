package main

import (
    "fmt"
    "math/rand"
    "time"
)

const (
    VARIABLE_COUNT = 20
    EXECUTIONS = 50
    MAX_DURATION_SEC = 10
    RANDOM_SWAP_PROB = 0.2
    CLAUSE_SIZE = 3
)

func randBool() bool {
    return (rand.Intn(422) < 211)
}

type Assignment []bool

func (assignment Assignment) swapVariable(variable int) {
    assignment[variable] = !assignment[variable]
}

func randomAssignment(variableCount int) Assignment {
    assignment := make(Assignment, variableCount, variableCount)
    for i := 0; i < variableCount; i++ {
        assignment[i] = randBool()
    }
    return assignment
}


type VariableAndNegated struct {
    variable int
    negated bool
}

type Clause []VariableAndNegated

func (clause Clause) IsTrue(assignment Assignment) bool {
    for _, varAndNeg := range clause {
        expected := !varAndNeg.negated
        if expected == assignment[varAndNeg.variable] {
            return true
        }
    }
    return false
}

// Reservoir algo
func randomClause(variableCount int) Clause {
    clause := make(Clause, CLAUSE_SIZE, CLAUSE_SIZE)
    for i := 0; i < CLAUSE_SIZE; i++ {
        clause[i] = VariableAndNegated{i, randBool()}
    }
    for i := CLAUSE_SIZE; i < variableCount; i++ {
        swapWith := rand.Intn(i)
        if swapWith < CLAUSE_SIZE {
            clause[swapWith].variable = i
        }
    }
    return clause
}


type Problem []Clause

func (problem Problem) randomFalseClause(assignment Assignment) int {
    falseClauses := make([]int, 0, len(problem))
    for i, clause := range problem {
        if !clause.IsTrue(assignment) {
            falseClauses = append(falseClauses, i)
        }
    }
    if len(falseClauses) == 0 {
        return -1
    } else {
        return falseClauses[rand.Intn(len(falseClauses))]
    }
}

func (problem Problem) randomVariable(clause int) int {
    variableAndNeg := rand.Intn(len(problem[clause]))
    return problem[clause][variableAndNeg].variable
}

func (problem Problem) bestVariable(clause int, assignment Assignment) int {
    bestVariable, maxTrueCount := -1, -1
    for _, varAndNeg := range problem[clause] {
        v := varAndNeg.variable
        assignment[v] = !assignment[v]
        trueCount := problem.trueCount(assignment)
        assignment[v] = !assignment[v]

        if bestVariable == -1 || trueCount > maxTrueCount {
            bestVariable = v
            maxTrueCount = trueCount
        }
    }
    return bestVariable
}

func (problem Problem) trueCount(assignment Assignment) int {
    count := 0
    for _, clause := range problem {
        if clause.IsTrue(assignment) {
            count++
        }
    }
    return count
}

func randomProblem(variableCount, clauseCount int) Problem {
    problem := make(Problem, clauseCount, clauseCount)
    for i := 0; i < clauseCount; i++ {
        problem[i] = randomClause(variableCount)
        // TODO - avoid duplicate clauses?
        // Requires comparing to previous ones, probably normalizing.
    }
    return problem
}


func singleSolveRandom(variableCount, clauseCount int) (bool, float64) {
    problem := randomProblem(variableCount, clauseCount)
    assignment := randomAssignment(variableCount)
    terminated := false

    start := time.Now()

    for time.Since(start).Seconds() < MAX_DURATION_SEC && !terminated {
        clause := problem.randomFalseClause(assignment)
        if clause == -1 {
            terminated = true
        } else {    
            var variableToSwap int
            if rand.Float64() < RANDOM_SWAP_PROB {
                variableToSwap = problem.randomVariable(clause)
            } else {
                variableToSwap = problem.bestVariable(clause, assignment)
            }
            assignment.swapVariable(variableToSwap)
        }
    }

    elapsedSec := time.Since(start).Seconds()
    return terminated, float64(elapsedSec)
}

func walksatSolveRandom(variableCount, clauseCount, executions int) (int, float64) {
    terminations := 0
    execTimeSum := 0.0

    for i := 0; i < executions; i++ {
        if i % 10 == 0 {
            fmt.Printf("  ...execution #%d\n", i)
        }
        terminated, execTime := singleSolveRandom(variableCount, clauseCount)
        if terminated {
            terminations++
            execTimeSum += execTime
        }
    }

    if terminations == 0 {
        return 0, -1.0
    } else {
        return terminations, execTimeSum / float64(terminations)
    }
}

func main() {
    rand.Seed(422)

    // for ratio := 1; ratio <= 10; ratio++ {
    for clauseCount := 70; clauseCount <= 105; clauseCount += 5 {
        // clauseCount := ratio * VARIABLE_COUNT
        fmt.Printf("Clause count = %d\n", clauseCount)
        terminations, execTime := walksatSolveRandom(VARIABLE_COUNT, clauseCount, EXECUTIONS)
        fmt.Printf("%d\t%d\t%.8f ms\n", clauseCount, terminations, execTime * 1000)
    }
}

// Debugging 
func (assignment Assignment) String() string {
    result := "{"
    for i, v := range assignment {
        if i > 0 {
            result += ","
        }
        if v {
            result += fmt.Sprintf("%d = T", i)
        } else {
            result += fmt.Sprintf("%d = F", i)
        }
    }
    return result + "}"
}
func (problem Problem) String() string {
    result := problem[0].String()
    for i := 1; i < len(problem); i++ {
        result += " && " + problem[i].String()
    }
    return result
}
func (clause Clause) String() string {
    result := "(" + clause[0].String()
    for i := 1; i < len(clause); i++ {
        result += " || " + clause[i].String()
    }
    return result + ")"
}
func (varAndNeg VariableAndNegated) String() string {
    if varAndNeg.negated {
        return fmt.Sprintf("!%d", varAndNeg.variable)
    } else {
        return fmt.Sprintf("%d", varAndNeg.variable)
    }
}
