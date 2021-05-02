package user

import "github.com/coolprototype-9000/cp9-proto/client"

func bind(args ...string) {
	if len(args) != 4 {
		printf("usage: bind <name> <old> <a|b|r>")
		return
	}

	var bt client.BindType
	switch args[3] {
	case "a":
		bt = client.After
	case "b":
		bt = client.Before
	case "r":
		bt = client.Replace
	default:
		printf("usage: bind <name> <old> <a|b|r>")
		return
	}

	st := p.Bind(args[1], args[2], bt)
	if st < 0 {
		printf("failed to bind: %s\n", p.Errstr())
	}
}
