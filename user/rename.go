package user

func rename(args ...string) {
	if len(args) != 3 {
		Printf("usage: rename <old> <new>\n")
		return
	}

	if p.Rename(args[1], args[2]) < 0 {
		Printf("Error: %s\n", p.Errstr())
	}
}
