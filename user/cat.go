package user

import "github.com/coolprototype-9000/cp9-proto/nine"

func cat(args ...string) {
	if len(args) == 1 {
		printf("unsupported, web terminal can't handle eofs afaik\n")
	} else {
		for i := 1; i < len(args); i++ {
			if doCat(args[i]) < 0 {
				return
			}
		}
	}
}

func doCat(fname string) int {
	fd := p.Open(fname, nine.OREAD)
	if fd < 0 {
		printf("cat: %s: %s\n", fname, p.Errstr())
		return -1
	}

	printf(p.Read(fd, maxQt))
	p.Close(fd)
	return 0
}
