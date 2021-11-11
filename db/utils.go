package db

func max(a, b uint) uint {
	if a > b {
		return a
	}
	return b
}

func ifZero(a, b uint) uint {
	if a > 0 {
		return a
	}
	return b
}
