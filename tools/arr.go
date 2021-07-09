package tools

func Includes(arr []string, target string) bool {
	for _, s := range arr {
		if target == s {
			return true
		}
	}
	return false
}

func Reverse(arr []interface{}) []interface{} {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		temp := arr[length-1-i]
		arr[length-1-i] = arr[i]
		arr[i] = temp
	}
	return arr
}
