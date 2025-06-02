package stringx

import "strings"

var (
	ConsoleColorReset    = "\033[0m"
	ConsoleColorFgRed    = "\033[31m"
	ConsoleColorFgGreen  = "\033[32m"
	ConsoleColorFgYellow = "\033[33m"
	ConsoleColorFgBlue   = "\033[34m"
	ConsoleColorFgPurple = "\033[35m"
	ConsoleColorFgCyan   = "\033[36m"
	ConsoleColorFgGray   = "\033[37m"

	ConsoleColorBgRed    = "\033[41m"
	ConsoleColorBgGreen  = "\033[42m"
	ConsoleColorBgYellow = "\033[43m"
	ConsoleColorBgBlue   = "\033[44m"
	ConsoleColorBgPurple = "\033[45m"
	ConsoleColorBgCyan   = "\033[46m"
	ConsoleColorBgGray   = "\033[47m"
)

func FormatColored(color string, text string) string {
	return color + text + ConsoleColorReset
}

func FormatColoredCenter(color string, text string, length int) string {
	if length < len([]rune(text)) {
		return color + text + ConsoleColorReset
	}
	padding := (length - len([]rune(text))) / 2
	left := padding + (length-len([]rune(text)))%2
	right := padding
	return color + strings.Repeat(" ", left) + text + strings.Repeat(" ", right) + ConsoleColorReset
}

func FormatColoredRight(color string, text string, length int) string {
	if length < len([]rune(text)) {
		return color + text + ConsoleColorReset
	}
	padding := length - len([]rune(text))
	return color + strings.Repeat(" ", padding) + text + ConsoleColorReset
}
