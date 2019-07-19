package slice

//Contains exported
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Diff exported
func Diff(a []string, b []string) []string {
	var c []string
	for _, e := range a {
		if !Contains(b, e) {
			c = append(c, e)
		}
	}
	return c
}
