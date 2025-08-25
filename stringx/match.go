package stringx

func MatchPartial(s, by string) int {
	matchedChars := 0
	for i := 0; i < len(s); i++ {
		if matchedChars == 0 {
			if s[i] == by[0] {
				matchedChars++
			}
		} else {
			if s[i] != by[matchedChars] {
				return matchedChars
			}
			matchedChars++
			if matchedChars == len(by) {
				break
			}
		}
	}
	return matchedChars
}
