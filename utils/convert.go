package utils

import "strconv"

func StringToFloat64(s string) (float64, error) {
	t, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1, err
	}

	return t, nil
}
