package user

func unmount(args ...string) {
	if len(args) == 2 {
		printf("usage: bind <name> <old>")
		return
	}

	st := p.Unmount(args[1], args[2])
	if st < 0 {
		printf("failed to unmount: %s\n", p.Errstr())
	}
}
