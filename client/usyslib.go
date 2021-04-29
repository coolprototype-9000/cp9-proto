package client

import (
	"path"

	"github.com/coolprototype-9000/cp9-proto/nine"
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

func (p *Proc) Stat(name string) *nine.Stat {
	kc, err := p.evaluate(name, true)
	if err != nil {
		p.errstr = "failed to evaluate name, no such file/dir?"
		return nil
	}
	st, err := fStat(kc)
	if err != nil {
		p.errstr = err.Error()
		return nil
	}
	return &st
}

func (p *Proc) Fstat(fd int) *nine.Stat {
	// Can't just stat this path, because
	// the path may have changed beneath us
	kc, ok := p.fdTbl[fd]
	if !ok {
		p.errstr = "no such fd"
		return nil
	}
	st, err := fStat(kc)
	if err != nil {
		p.errstr = err.Error()
		return nil
	}
	return &st
}

func (p *Proc) Chdir(pathname string) int {
	ncwd, err := p.evaluate(pathname, true)
	if err != nil {
		p.errstr = "no such file or directory"
		return -1
	}
	if !kchanCmp(p.cwd, &rootChannel) {
		fClunk(p.cwd)
	}
	p.cwd = ncwd
	return 0
}
