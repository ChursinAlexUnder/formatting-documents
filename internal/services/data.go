package services

func InSlice(elem string, slice []string) bool {
	for _, str := range slice {
		if elem == str {
			return true
		}
	}
	return false
}

func Add(a, b int) int {
	return a + b
}

func AllTrue(arr []bool) bool {
	for _, v := range arr {
		if !v {
			return false
		}
	}
	return true
}
