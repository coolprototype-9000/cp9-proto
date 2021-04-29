package client

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

type BindType byte

const (
	Replace BindType = iota
	Before
	After
)

func (p *Proc) Bind(name string, old string, mode BindType) int {
	var newc *kchan

	// Check if the old name is special, we could be bootstrapping
	tp, err := parseDevMnt(name)
	if err == nil {
		// Dial localhost using hardcoded ports
		var lp uint16
		switch tp {
		case nine.DevCons:
			lp = 5640
		case nine.DevRamFs:
			lp = 5641
		default:
			lp = 5642
		}

		c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", lp))
		if err != nil {
			log.Fatalf("No local file service over port %d", lp)
		}
		bskc := &kchan{
			c: &c,
		}
		newc, err = fVersion(bskc, 0, nine.NineVersion)
		if err != nil {
			log.Fatalf("Local file service did not respond")
		}
		err = fAttach(newc, mkFid(), p.owner, old)
		if err != nil {
			log.Fatalf("Local file server refused to attach")
		}

	} else {
		newc, err = p.evaluate(name, false)
		if err != nil {
			p.errstr = "failed to evaluate old path"
			return -1
		}
	}

	// newc is set up. now, evaluate the old path
	if _, err := parseDevMnt(old); err == nil {
		p.errstr = "can't use that old name!"
		return -1
	}
	oldc, err := p.evaluate(old, true)
	if err != nil {
		p.errstr = "failed to evaluate new path"
		return -1
	}

	switch mode {
	case Replace:
		// bind oldc->nc first
		p.mnt.bind(oldc, newc, true)
	case After:
		// bind oldc->oldc, THEN oldc->newc
		p.mnt.bind(oldc, oldc, true)
		p.mnt.bind(oldc, newc, false)
	case Before:
		// bind oldc->newc, THEN oldc->oldc
		p.mnt.bind(oldc, newc, true)
		p.mnt.bind(oldc, oldc, false)
	}
	return 0
}

func (p *Proc) Fd2Path(fd int) string {
	if nm, ok := p.fdTbl[fd]; ok {
		return nm.name
	}
	p.errstr = "no such file or directory"
	return ""
}

func (p *Proc) Unmount(name string, old string) int {
	var newc *kchan
	var oldc *kchan

	// Check whether we are unmounting a local device
	_, err := parseDevMnt(name)
	_, err2 := parseDevMnt(old)
	if err == nil || err2 != nil {
		p.errstr = "currently unsupported"
		return -1
	}

	// Evaluate oldc
	oldc, err = p.evaluate(old, true)
	if err != nil {
		p.errstr = "no such old name to unmount off"
		return -1
	}

	// Evaluate newc
	newc, err = p.evaluate(name, false)
	if err != nil {
		p.errstr = "no such name to unmount"
		return -1
	}

	// Okay, we need to make sure the user isn't
	// shooting themselves in the foot. If anybody
	// is using anything deriving from newc, fail
	// How do we check if newc has children?
	// We can abuse our invariant and check through
	// the fd table of our process to see if any lexical
	// descendents, INCLUDING CWD AND ROOT, are in use
	ss := newc.name
	if strings.Contains(p.cwd.name, ss) || strings.Contains(rootChannel.name, ss) {
		p.errstr = "device is busy"
		return -1
	} else {
		for _, kc := range p.fdTbl {
			if strings.Contains(kc.name, ss) {
				p.errstr = "device is busy"
				return -1
			}
		}
	}

	// newc has no children in use
	// we can unbind it now
	if p.mnt.unbind(oldc, newc) != nil {
		log.Fatal("internal inconsistency")
	}

	// next, check to see if there is anything
	// mapping old -> * where * != old
	// if there is NOT, blindly remove old -> old
	candidates := p.mnt.forwardEval(oldc)
	ununion := true
	for _, cand := range candidates {
		if !kchanCmp(oldc, cand) {
			ununion = false
			break
		}
	}
	if ununion {
		p.mnt.unbind(oldc, oldc)
	}

	// lastly, we need to figure out a way to gc
	// network connections. mount is not yet implemented
	// so no need!

	return 0
}
