package user

import "github.com/coolprototype-9000/cp9-proto/mr"

func mrw(args ...string) {
	if len(args) != 1 {
		Printf("Usage: mrw")
		return
	}

	Printf("Working...")

	mr.Worker(mr.DoMap, mr.DoReduce, p)
	Printf("Done\n")
}
