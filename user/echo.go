package user

import "strings"

func echo(args ...string) {
	printf(strings.Join(args[1:], " ") + "\n")
}
