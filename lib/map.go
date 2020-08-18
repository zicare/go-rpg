package lib

//MapFilterByValue exported
func MapFilterByValue(m *map[string]interface{}, val interface{}) *map[string]interface{} {

	for k, v := range *m {
		if v == val {
			delete(*m, k)
		}
	}
	return m
}
