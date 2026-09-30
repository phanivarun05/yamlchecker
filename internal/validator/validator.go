package validator

func CheckRequired(data map[string]interface{}, required []string) []string {
	var missing []string

	for _, key := range required {
		_, ok := data[key]
		if !ok {
			missing = append(missing, key)
		}
	}
	return missing
}
