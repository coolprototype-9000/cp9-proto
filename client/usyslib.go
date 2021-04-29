package client

import (
	"path"
)

func (p *Proc) Create(name string, mode byte, perm uint32) int {
	// Evaluate the path up to the last element
	if cleanPath(name) == "." || cleanPath(name) == "/" {
		p.errstr = "cannot create that file"
		return -1
	}
	initc, err := p.evaluate(path.Dir(name), false)
	if err != nil {
		p.errstr = "no such parent directory"
		return -1
	}

	nc, err := fWalk(initc, mkFid(), []string{})
	if err != nil {
		p.errstr = "failed to dup fd"
		return -1
	}
	newf := path.Base(name)
	if err := fCreate(nc, newf, perm, mode); err != nil {
		p.errstr = err.Error()
		return -1
	}
	if !kchanCmp(initc, &rootChannel) && !kchanCmp(initc, p.cwd) {
		fClunk(initc)
	}
	nf := p.mkFd()
	p.fdTbl[nf] = nc
	return nf
}
