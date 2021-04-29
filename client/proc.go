package client

// Collection of active files
type proc struct {
	mnt   mountTable
	cwd   *kchan
	maxfd uint64
	fdTbl map[uint64]kchan
}
