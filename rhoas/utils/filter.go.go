package utils

func Filter(items []map[string]interface{}, fields []string) []map[string]interface{} {
	answer := make([]map[string]interface{}, 0)
	for _, item := range items {
		filtered := make(map[string]interface{})
		for k, v := range item {
			keep := false
			for _, f := range fields {
				if f == k {
					keep = true
				}
			}
			if keep {
				filtered[k] = v
			}
		}
		answer = append(answer, filtered)
	}
	return answer
}
