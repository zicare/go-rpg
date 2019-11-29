package lib

//XOR exported
func XOR(vals ...interface{}) bool {

	var c = 0

	for _, val := range vals {

		switch v := val.(type) {
		case bool:
			if v == true {
				c++
			}
		default:
			if v != nil {
				c++
			}
		}
	}

	if c == 1 {
		return true
	}

	return false
}
