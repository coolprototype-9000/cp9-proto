package consfs

import "github.com/coolprototype-9000/cp9-proto/nine"

type fidDesc struct {
	fid      nine.Fid
	owner    string
	open     bool
	openMode byte
}

const rootId = 0
const listenId = 1

type ConsFs struct {
	devNumber uint32
	fidTable  map[uint64]map[fidDesc]uint64

	statTable map[uint64]nine.Stat
	fileTable map[uint64]*service
}
