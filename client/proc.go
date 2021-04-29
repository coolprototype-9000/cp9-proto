package client

// Collection of active files
type Proc struct {
	mnt    *mountTable
	owner  string
	cwd    *kchan
	maxfd  int
	fdTbl  map[int]*kchan
	errstr string
}

func (p *Proc) mkFd() int {
	p.maxfd++
	return p.maxfd - 1
}

func MkProc(cwd *kchan, owner string) *Proc {
	if cwd == nil {
		cwd = &rootChannel
	}
	return &Proc{
		mnt:   mkFreshMountTable(),
		cwd:   cwd,
		owner: owner,
		fdTbl: make(map[int]*kchan),
	}
}
