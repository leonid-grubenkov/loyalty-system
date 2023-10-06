package utils

import "strconv"

func ParseOrder(body string) int {

	num, err := strconv.Atoi(body)
	if err != nil {
		return -1
	}
	return num
}
