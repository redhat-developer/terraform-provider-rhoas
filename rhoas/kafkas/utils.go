package kafkas

func fixBootstrapServerHosts(items []map[string]interface{}) []map[string]interface{} {
	// Fix the bootstrap server host to snake case
	answer := make([]map[string]interface{}, 0)
	for _, entry := range items {
		entry["bootstrap_server"] = entry["bootstrapServerHost"]
		delete(entry, "bootstrapServerHost")
		answer = append(answer, entry)
	}
	return answer
}
