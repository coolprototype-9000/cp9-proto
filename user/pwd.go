package user

import "github.com/coolprototype-9000/cp9-proto/nine"

// p.Errstr() -> string error
// p.Fd2Path(fd) -> where a given file is
// "." = the current directory
func pwd(args ...string) {
	if len(args) != 1 {
		printf("too many arguments\n")
	}

	fd := p.Open(".", nine.OREAD)
	if fd < 0 {
		printf("pwd: %s: %s\n", args[1], p.Errstr())
		return
	}
	res := p.Fd2Path(fd)
	printf(res + "\n")
	p.Close(fd)
}
