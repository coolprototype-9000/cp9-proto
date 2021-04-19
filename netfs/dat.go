package netfs

import "github.com/coolprototype-9000/cp9-proto/nine"

type fidDesc struct {
	fid      nine.Fid
	owner    string
	open     bool
	openMode byte
}

type NetFs struct {
	devNumber uint32
	fidTable  map[uint64]map[fidDesc]uint64

	rootstat  *nine.Stat
	clonestat *nine.Stat
	cons      map[uint64]*netInst
}
