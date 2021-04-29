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

	newf := path.Base(name)
	if fCreate(initc, newf, perm, mode) != nil {
		p.errstr = err.Error()
		return -1
	}
	return 0
}
