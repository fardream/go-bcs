package bcs_test

func sliceEqual[E1 ~[]T, E2 ~[]T, T comparable](s1 E1, s2 E2) bool {
	if len(s1) != len(s2) {
		return false
	}
	l := len(s1)
	for i := 0; i < l; i++ {
		if s1[i] != s1[i] {
			return false
		}
	}

	return true
}
