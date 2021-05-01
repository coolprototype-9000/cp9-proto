package client

import (
	"bufio"
	"fmt"
	"os"
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
	p.fdTbl[nf] = make([]*kchan, 1)
	p.fdTbl[nf][0] = nc
	p.seekTbl[nf] = 0
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

	if p.isSpecialFd(fd) {
		p.errstr = "can't stat stdout/stderr/stdin"
		return nil
	}

	st, err := fStat(kc[0])
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

	if p.isSpecialFd(fd) {
		goto donenet
	}

	// We have an invariant that created/opened
	// fds are dupped, so deleting blindly is fine
	// and won't impact the mount table. The exception
	// to this invariant is currently the std*s, but
	// these don't hit the mount table.
	for _, ent := range kc {
		fClunk(ent)
	}

donenet:
	delete(p.fdTbl, fd)
	delete(p.seekTbl, fd)
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

func (p *Proc) Open(file string, mode byte) int {
	// Head to the file, but don't make the final deref
	// in case this is a union directory or union file
	kcs := make([]*kchan, 1)
	kc, err := p.evaluate(file, true)
	if err != nil {
		p.errstr = "no such file or directory"
		return -1
	}
	kcs[0] = kc

	// Stat
	st, err := fStat(kc)
	if err != nil {
		p.errstr = "failed to stat file"
		return -1
	}

	if st.Q.Flags&nine.FDir > 0 {
		// Run the last forward eval ourselves
		kcl := p.mnt.forwardEval(kc)
		if len(kcl) > 0 {
			kcs = kcl
			for i, kc := range kcs {
				kcs[i], err = fWalk(kc, mkFid(), []string{})
				if err != nil {
					p.errstr = "failed to dup fid for preopen"
					return -1
				}
			}
		}
	} else {
		kcl := p.mnt.forwardEval(kc)
		if len(kcl) > 0 {
			kcs = []*kchan{kcl[0]}
			kcs[0], err = fWalk(kc, mkFid(), []string{})
			if err != nil {
				p.errstr = "failed to dup fid for preopen"
				return -1
			}
		}
	}

	ncs := make([]*kchan, 0)
	for _, kc := range kcs {
		nc, err := fWalk(kc, mkFid(), []string{})
		if err != nil {
			p.errstr = "failed to dup fid for open"
			return -1
		}

		if err := fOpen(nc, mode); err != nil {
			p.errstr = err.Error()
			return -1
		}

		if !kchanCmp(kc, &rootChannel) && !kchanCmp(kc, p.cwd) {
			fClunk(kc)
		}

		ncs = append(ncs, nc)

	}
	nf := p.mkFd()
	p.fdTbl[nf] = ncs
	p.seekTbl[nf] = 0
	fmt.Printf("length ncs: %d\n", len(ncs))
	return nf
}

func (p *Proc) Read(fd int, count uint32) string {
	kcs, ok := p.fdTbl[fd]
	if !ok {
		p.errstr = "no such fd"
		return ""
	}

	if p.isSpecialFd(fd) {
		switch kcs[0].name {
		case "STDIN":
			rdr := bufio.NewReader(os.Stdin)
			txt, _ := rdr.ReadString('\n')
			if len(txt) > int(count) {
				txt = txt[:count]
			}
			return txt
		case "STDOUT":
			p.errstr = "can't read from stdout"
			return ""
		case "STDERR":
			p.errstr = "can't read from stderr"
			return ""
		}
	}

	if len(kcs) == 1 {
		kc := kcs[0]

		b, err := fRead(kc, p.seekTbl[fd], count)
		if err != nil {
			p.errstr = err.Error()
			return ""
		}
		p.seekTbl[fd] += uint64(len(b))
		return string(b)
	}

	// Union directory
	b := make([]byte, 0)
	icnt := int64(count)

	for _, kc := range kcs {
		nb, err := fRead(kc, p.seekTbl[fd], ^uint32(0))
		if err != nil {
			p.errstr = err.Error()
			return ""
		}

		// This is hacky but its fine
		// All or nothing union reads
		icnt -= int64(len(nb))
		if icnt < 0 {
			p.errstr = "insufficient room in buffer (HACKY, SEE CODE)"
			return ""
		}
		b = append(b, nb...)
	}

	p.seekTbl[fd] += uint64(len(b))
	return string(b)
}

func (p *Proc) Write(fd int, data string) int {
	kcs, ok := p.fdTbl[fd]
	if !ok {
		p.errstr = "no such fd"
		return -1
	}

	if p.isSpecialFd(fd) {
		switch kcs[0].name {
		case "STDIN":
			p.errstr = "can't write to stdin"
			return 0
		case "STDOUT":
			fmt.Printf("%s", data)
			return len(data)
		case "STDERR":
			fmt.Fprintf(os.Stderr, "%s", data)
			return len(data)
		}
	}

	cnt, err := fWrite(kcs[0], p.seekTbl[fd], data)
	if err != nil {
		p.errstr = err.Error()
		return -1
	}
	p.seekTbl[fd] += uint64(cnt)
	return int(cnt)
}
