package utils

import "strconv"

func ParseInt(val string) int {
	valInt, _ := strconv.Atoi(val)
	return valInt
}
