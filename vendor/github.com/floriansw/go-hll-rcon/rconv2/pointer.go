package rconv2

func toInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
