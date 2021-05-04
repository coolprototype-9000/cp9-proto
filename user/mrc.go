package user

import (
	"fmt"
	"time"

	"github.com/coolprototype-9000/cp9-proto/mr"
)

func mrc(args ...string) {

	Printf("Working")

	base := "sherlock-holmes-a"
	names := []string{}
	for _, c := range "abcdefghijklmn" {
		names = append(names, fmt.Sprintf("%s%v", base, c))
	}

	m := mr.MakeCoordinator(names, 10, p)
	for !m.Done() {
		Printf(".")
		time.Sleep(time.Second)
	}

	Printf("Done!\n")

	time.Sleep(time.Second)
}
