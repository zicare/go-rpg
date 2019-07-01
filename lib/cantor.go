package lib

//Cantor exported
func Cantor(a, b int64) int64 {
	return ((a+b)*(a+b+1) + b) / 2
}
