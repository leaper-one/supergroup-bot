package tools

func Includes(arr []string, target string) bool {
	for _, s := range arr {
		if target == s {
			return true
		}
	}
	return false
}
