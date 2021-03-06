package user

func cd(args ...string) {
	if len(args) == 1 {
		doChdir("/")
	} else if len(args) > 2 {
		printf("too many arguments\n")
	} else {
		doChdir(args[1])
	}
}

func doChdir(where string) {
	st := p.Chdir(where)
	if st < 0 {
		printf("failed to chdir: %s\n", p.Errstr())
	}
}
