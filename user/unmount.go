package user

func unmount(args ...string) {
	if len(args) != 3 {
		Printf("usage: unmount <name> <old>")
		return
	}

	st := p.Unmount(args[1], args[2])
	if st < 0 {
		Printf("failed to unmount: %s\n", p.Errstr())
	}
}
