package user

import (
	"fmt"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

const maxQt = ^uint32(0)

func Printf(format string, param ...interface{}) {
	str := fmt.Sprintf(format, param...)
	p.Write(1, str)
}

func Dirread(fd int) []*nine.Stat {
	b := []byte(p.Read(fd, maxQt))

	stats := make([]*nine.Stat, 0)
	for len(b) > 0 {
		stat, nb := nine.UnmarshalStat(b)
		b = nb
		stats = append(stats, &stat)
	}
	return stats
}
