package lib

import "time"

//TimeRange exported
type TimeRange struct {
	From, To time.Time
}

//Pair exported
type Pair struct {
	A, B interface{}
}

//Triplet exported
type Triplet struct {
	A, B, C interface{}
}

//Quartet exported
type Quartet struct {
	A, B, C, D interface{}
}
