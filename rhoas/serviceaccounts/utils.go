package serviceaccounts

func fixClientIDAndClientSecret(items []map[string]interface{}, existingClientSecret *string) []map[string]interface{} {
	// Fix the client id and client secret
	answer := make([]map[string]interface{}, 0)
	for _, entry := range items {
		if existingClientSecret != nil {
			entry["client_secret"] = existingClientSecret
		}
		answer = append(answer, entry)
	}
	return answer
}
