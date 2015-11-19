package main

import (
    "fmt"
)

type Distribution []float64
type Model []Distribution
type TransitionModel Model
type SensorModel Model

func main() {
    // Assignment example:
    T := TransitionModel{
        {0.7, 0.3},
        {0.4, 0.6},
    }
    S := SensorModel{
        {0.8, 0.2},
        {0.3, 0.7},
    }
    e := []int{asState(true), asState(false), asState(true)}

    // Umbrella example:
    /*
    T := TransitionModel{
        {0.7, 0.3},
        {0.3, 0.7},
    }
    S := SensorModel{
        {0.9, 0.1},
        {0.2, 0.8},
    }
    e := []int{asState(true), asState(true), asState(false), asState(true), asState(true)}; 
    */
    // estimate(0, e, T, S)
    // for i := 0; i <= 5; i++ {
        // fmt.Printf("Estimate for %d: %s\n", i, estimate(i, e, T, S))
    // } 

    // Wikipedia example:
    /*
    T := TransitionModel {
        {0.7, 0.3},
        {0.4, 0.6},
    }
    S := SensorModel {
        {0.5, 0.4, 0.1},
        {0.1, 0.3, 0.6},
    }
    e := []int{0, 1, 2}
    */

    mostLikely := viterbi(e, T, S)
    fmt.Printf("Most likely states: ")
    for _, v := range mostLikely {
        fmt.Printf("%d -> ", v)
    }
    fmt.Printf("end.\n")
    
    // filter(e, T, S)
    // backFilter(e, T, S)
    // for i := 1; i <= 3; i++ {
         // estimate(i, e, T, S)
    // } 
}

func viterbi(observed []int, T TransitionModel, S SensorModel) []int {
    steps := len(observed)
    states := len(T)
    initial := flatDistribution()

    values, prevState := make([]Distribution, steps, steps), make([][]int, steps, steps)
    for i := 0; i < steps; i++ {
        values[i], prevState[i] = make(Distribution, states, states), make([]int, states, states)
    }

    for step := 0; step < steps; step++ {
        for state := 0; state < states; state++ {
            if step == 0 {
                values[step][state] = initial[state]
            } else {
                maxValue := 0.0
                for prev := 0; prev < states; prev++ {
                    next := values[step - 1][prev] * T[prev][state]
                    if next > maxValue {
                        maxValue = next
                        prevState[step][state] = prev
                    }
                }
                values[step][state] = maxValue
            }
            values[step][state] *= S[state][observed[step]] 
        }
        values[step].normalize()
    }

    for state := 0; state < states; state++ {
        for step := 0; step < steps; step++ {
            fmt.Printf("%.3f\t", values[step][state])
        }
        fmt.Printf("\n")
    }

    for state := 0; state < states; state++ {
        for step := 0; step < steps; step++ {
            fmt.Printf("%d\t", prevState[step][state])
        }
        fmt.Printf("\n")
    }

    bestPath := make([]int, steps, steps)
    bestPath[steps - 1] = maxIdx(values[steps - 1])
    for at := steps - 1; at > 0; at-- {
        bestPath[at - 1] = prevState[at][bestPath[at]]
    }
    return bestPath
}

// Convert state value (T/F) into an index used for the matrices.
func asState(value bool) int {
    // Out of convention, all matrices have [true, true] in the top left corner (0, 0).
    if value {
        return 0
    } else {
        return 1
    }
}

// NOTE: at is ONE-BASED!
func estimate(at int, observed []int, T TransitionModel, S SensorModel) Distribution {
    // Using current value:
    dForward, dBackward := flatDistribution(), flatDistribution()

    if at > 0 {
        dForward = filter(observed[:at], T, S)
    }
    if at < len(observed) {
        dBackward = backFilter(observed[at:], T, S)
    }

    d := make(Distribution, 2, 2)
    for i := 0; i < 2; i++ {
        d[i] = dForward[i] * dBackward[i]
    }
    d.normalize()
    fmt.Printf("%s . %s => %s\n", dForward, dBackward, d)
    return d
}

func filter(observed []int, T TransitionModel, S SensorModel) Distribution {
    d := flatDistribution()
    invS := transpose(Model(S))
    invT := transpose(Model(T))

    // fmt.Printf("start: %s\n", d)
    for _, e := range observed {
        // fmt.Printf("%s -> ", d)
        d.timesModel(invT)
        // fmt.Printf("%s -> ", d)
        d.timesDistribution(invS[e])
        d.normalize()
        // fmt.Printf("%s\n", d)
    }
    return d
}

func backFilter(observed []int, T TransitionModel, S SensorModel) Distribution {
    rObs := reverseArray(observed)

    d := flatDistribution()
    invS := transpose(Model(S))
    // invT := transpose(Model(T))

    // fmt.Printf("BACK: %s\n", d)
    for _, e := range rObs {
        // fmt.Printf("%s -> ", d)
        // d.normalize()
        d.timesDistribution(invS[e])
        // fmt.Printf("%s -> ", d)
        d.timesModel(Model(T))
        d.normalize()
        // fmt.Printf("%s\n", d)
    }
    return d
}


func (d Distribution) timesModel(m Model) {
    // fmt.Printf("\n%s * \n", d)
    size := len(d)
    result := make(Distribution, size, size)
    // for i := 0; i < size; i++ {
        // fmt.Printf("d[%d] = %s . %s\n", i, m[i], d)
    // }
    for i := 0; i < size; i++ { 
        result[i] = dot(m[i], d)
    }
    for i := 0; i < size; i++ {
        d[i] = result[i]
    }
}

func (d Distribution) timesDistribution(other Distribution) {
    for i := range d {
        d[i] *= other[i]
    }
}

func dot(d1, d2 Distribution) float64 {
    if len(d1) != len(d2) {
        panic("Vector dot product requires two inputs of same dimension")
    }
    sum := 0.0
    for i := range d1 {
        sum += d1[i] * d2[i]
    }
    return sum
}

func (d Distribution) clone() Distribution {
    result := make(Distribution, len(d), len(d))
    for i := range d {
        result[i] = d[i]
    }
    return result
}

// Given a forwards transition model, reverse it into a backwards one.
func reverseModel(model Model) []Distribution {
    t := transpose(model)
    t.normalize()
    return t
}

// Reverse an array
func reverseArray(in []int) []int {
    size := len(in)
    out := make([]int, size, size)
    for i, v := range in {
        out[size - 1 - i] = v
    }
    return out
}

// Transpose, return a new matrix
func transpose(m Model) Model {
    rows, cols := len(m), len(m[0])
    if rows != cols {
        panic("Transpose requires square matrix!")
    }

    result := make([]Distribution, rows, rows)
    for r := 0; r < rows; r++ {
        result[r] = make(Distribution, cols, cols)
    }

    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            result[r][c] = m[c][r]
        }
    }
    return result;
}

// Normalize each row inline.
func (m Model) normalize() {
    for i := range m {
        m[i].normalize()
    }
}

// Normalize a single distribution.
func (r Distribution) normalize() {
    sum := 0.0
    for _, v := range r {
        sum += v
    }
    for i, _ := range r {
        r[i] /= sum
    }
}

// Distribution assuming all states are equally likely
func flatDistribution() Distribution {
    return Distribution{0.5, 0.5}
}

// Print Distribution
func (d Distribution) String() string {
    return fmt.Sprintf("<%.4f, %.4f>", d[0], d[1])
}

// Index of maximum value in slice
func maxIdx(v []float64) int {
    m, s := 0, len(v)
    for i := 1; i < s; i++ {
        if v[i] > v[m] {
            m = i
        }
    }
    return m
}
