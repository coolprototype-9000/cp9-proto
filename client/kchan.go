package client

import (
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Immutable data type
type kchan struct {
	name    string
	phyName string
	c       *net.Conn
	fid     nine.Fid
}

var rootChannel kchan = kchan{
	name: "/",
}

// Two kchans are the "same" if the names
// and networks are identical.
// Since kchans use full lexical names, this
// works for bind mounts.
func kchanCmp(a *kchan, b *kchan) bool {
	if a.phyName == b.phyName && a.c == b.c {
		return true
	}
	return false
}
