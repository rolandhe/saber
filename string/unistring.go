package string

func GetRuneLenOfString(s string) int {
	runes := []rune(s)
	return len(runes)
}
