package utils

import (
	"strconv"
)

func ParseOrder(body string) int {

	num, err := strconv.Atoi(body)
	if err != nil {
		return -1
	}

	if !isValidLuhn(body) {
		return -1
	}

	return num
}

func isValidLuhn(s string) bool {
	sum := 0
	alternate := false

	for i := len(s) - 1; i >= 0; i-- {
		digit := int(s[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
