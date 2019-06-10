package lib

//Max exported
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

//Min exported
func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
