package user

import "github.com/coolprototype-9000/cp9-proto/nine"

func mkdir(args ...string) {
	if len(args) == 1 {
		printf("too few arguments\n")
	}

	baseFlags := nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGW | nine.PGX | (nine.FDir << nine.FStatOffset)

	for i := 1; i < len(args); i++ {
		fd := p.Create(args[i], nine.OREAD, uint32(baseFlags))

		if fd < 0 {
			printf("mkdir: error: %s\n", p.Errstr())
			break
		}

	}

}
