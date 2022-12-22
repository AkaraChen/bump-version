package util

func FindIndex(array []string, item string) int {
	for index, value := range array {
		if value == item {
			return index
		}
	}
	return -1
}
