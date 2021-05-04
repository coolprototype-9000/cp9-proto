package user

import "github.com/coolprototype-9000/cp9-proto/nine"

func ls(argv ...string) {
	if len(argv) < 2 {
		do(".")
		return
	}
	for i := 1; i < len(argv); i++ {
		res := do(argv[i])
		if res < 0 {
			return
		}
	}
}

func do(f string) int {
	fd := p.Open(f, nine.OREAD)
	if fd < 0 {
		Printf("ls: open: %s\n", p.Errstr())
		return -1
	}

	st := p.Fstat(fd)
	if st == nil {
		Printf("ls: stat: %s\n", p.Errstr())
		p.Close(fd)
		return -1
	}

	if st.Q.Flags&nine.FDir == 0 {
		Printf("%s\t%s\t%d\t%d\n", st.Name, "fil", st.Size, st.Mode)
		return 0
	} else {
		stats := dirread(fd)
		for _, st := range stats {
			disp := "fil"
			if st.Q.Flags&nine.FDir > 0 {
				disp = "dir"
			}
			Printf("%s\t%s\t%d\t%d\n", st.Name, disp, st.Size, st.Mode)
		}
	}
	p.Close(fd)
	return 0
}
