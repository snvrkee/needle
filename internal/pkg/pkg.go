package pkg

func ShortString(str string, length int) string {
	runeStr := []rune(str)
	if length < len(runeStr) {
		return string(runeStr[:length-1])
	}
	return str
}

func IsAlphaString(str string) bool {
	for _, b := range str {
		if 'a' <= b && b <= 'z' ||
			'A' <= b && b <= 'Z' ||
			b == '_' {
			continue
		}
		return false
	}
	return true
}

// Catch returns (O, nil) if there is no panic; else (nil, E)
func Catch[I, O, E any](f func(I) O, input I) (output O, exception E) {
	defer func() {
		if r := recover(); r != nil {
			if exc, ok := r.(E); ok {
				exception = exc
				return
			}
			panic(r)
		}
	}()
	output = f(input)
	return
}
