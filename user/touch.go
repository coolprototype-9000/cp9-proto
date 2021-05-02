package user

import "github.com/coolprototype-9000/cp9-proto/nine"

func touch(args ...string) {
	if len(args) == 1 {
		printf("not enough arguments")
		return
	}

	baseFlags := nine.PUR | nine.PUW | nine.PGR | nine.PGW
	for i := 1; i < len(args); i++ {
		fd := p.Open(args[i], nine.OREAD)
		if fd < 0 {
			fd := p.Create(args[i], nine.OREAD, uint32(baseFlags))
			if fd < 0 {
				printf("failed to create file: %s\n", p.Errstr())
				break
			}
		}
	}
}
