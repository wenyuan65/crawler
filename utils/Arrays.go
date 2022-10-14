package utils

func Distinct[T comparable](array ...[]T) (result []T) {
	result = []T{}

	filter := make(map[T]int8)
	for i := 0; i < len(array); i++ {
		for _, url := range array[i] {
			if _, ok := filter[url]; !ok {
				filter[url] = 1
				result = append(result, url)
			}
		}
	}

	return
}
