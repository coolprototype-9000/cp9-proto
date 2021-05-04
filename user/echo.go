package user

import "strings"

func echo(args ...string) {
	Printf(strings.Join(args[1:], " ") + "\n")
}
