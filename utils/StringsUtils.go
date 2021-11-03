package utils

import "strings"

func GetStringInBetween(str string, startS string, endS string) (result string) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result
	}
	result = newS[:e]
	return result
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func RemoveElement(slice []string, remove string) []string {
	var i int
	for i = 0; i < len(slice); i++ {
		if slice[i] == remove {
			break
		}
	}

	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}
