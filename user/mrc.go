package user

import (
	"os"
	"time"

	"github.com/coolprototype-9000/cp9-proto/mr"
)

func mrc(args ...string) {
	if len(args) < 2 {
		Printf("Usage: mrc inputfiles...\n")
		return
	}

	Printf("Working")

	m := mr.MakeCoordinator(os.Args[1:], 10, p)
	for !m.Done() {
		Printf(".")
		time.Sleep(time.Second)
	}

	Printf("Done!\n")

	time.Sleep(time.Second)
}
