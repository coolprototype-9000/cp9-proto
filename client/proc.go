package client

// Collection of active files
type Proc struct {
	mnt     *mountTable
	owner   string
	cwd     *kchan
	maxfd   int
	fdTbl   map[int][]*kchan
	seekTbl map[int]uint64
	errstr  string
}

func (p *Proc) mkFd() int {
	p.maxfd++
	return p.maxfd - 1
}

func (p *Proc) isSpecialFd(fd int) bool {
	n := p.fdTbl[fd][0]
	switch n.name {
	case "STDOUT":
		return true
	case "STDIN":
		return true
	case "STDERR":
		return true
	default:
		return false
	}
}

func MkProc(cwd *kchan, owner string) *Proc {
	if cwd == nil {
		cwd = &rootChannel
	}

	nfdtbl := make(map[int][]*kchan)
	nfdtbl[0] = []*kchan{{name: "STDIN"}}
	nfdtbl[1] = []*kchan{{name: "STDOUT"}}
	nfdtbl[2] = []*kchan{{name: "STDERR"}}

	return &Proc{
		mnt:     mkFreshMountTable(),
		cwd:     cwd,
		owner:   owner,
		maxfd:   3,
		fdTbl:   nfdtbl,
		seekTbl: make(map[int]uint64),
	}
}
