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

func (p *Proc) Close(fd int) int {
	kc, ok := p.fdTbl[fd]
	if !ok {
		p.errstr = "no such fd"
		return -1
	}

	// We have an invariant that created/opened
	// fds are dupped, so deleting blindly is fine
	// and won't impact the mount table. The exception
	// to this invariant is currently the std*s, but
	// these don't hit the mount table.
	fClunk(kc)
	delete(p.fdTbl, fd)
	return 0
}

func (p *Proc) Remove(file string) int {
	kc, err := p.evaluate(file, false)
	if err != nil {
		p.errstr = "no such file or directory"
		return -1
	}

	// This'll work no matter what -- "phase errors"
	// However, if the entry is in the bind table, we'll balk
	// to prevent weird situations
	// We'll also prevent if you from removing over yourself
	// The server should prevent removal of populated dirs
	for _, mpr := range p.mnt.tbl {
		if kchanCmp(mpr.from, kc) || kchanCmp(mpr.to, kc) {
			p.errstr = "file is busy / mounted"
			return -1
		}
	}

	ss := kc.name
	if len(p.cwd.name) >= len(ss) && p.cwd.name[:len(ss)] == ss || ss == "/" {
		p.errstr = "you are using this file as an ancestor of/as your cwd"
		return -1
	}

	if fRemove(kc) != nil {
		p.errstr = err.Error()
		return -1
	}
	return 0
}
