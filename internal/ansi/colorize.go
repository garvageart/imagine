package ansi

import (
	"fmt"
	"strconv"
)

func Colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, Reset)
}
