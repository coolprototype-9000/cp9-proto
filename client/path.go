package client

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// This legitimately brought a smile to my face
// Thanks, plan 9 / golang crossover
func cleanPath(orig string) string {
	return path.Clean(orig)
}

// Walk a fresh fid to the pth in question,
// returning a fresh kchan. This kchan should
// be garbage collected if it is not active
// post evaluation. The path does NOT have
// to be clean, we do that just in case.
func (p *Proc) evaluate(pth string, estop bool) (*kchan, error) {
	els := strings.Split(cleanPath(pth), "/")
	cl := []*kchan{&rootChannel}

	if pth == "." {
		if estop {
			return p.cwd, nil
		} else {
			cl = p.mnt.forwardEval(p.cwd)
			if len(cl) > 0 {
				return cl[0], nil
			}
			return p.cwd, nil
		}
	} else if pth == "/" {
		fmt.Println("HERE")
		if estop {
			return &rootChannel, nil
		} else {
			cl = p.mnt.forwardEval(&rootChannel)
			if len(cl) > 0 {
				return cl[0], nil
			}
			return &rootChannel, nil
		}
	} else if pth[0] != '/' {
		if ncl := p.mnt.forwardEval(p.cwd); len(ncl) != 0 {
			cl = ncl
		}
	} else {
		if ncl := p.mnt.forwardEval(&rootChannel); len(ncl) != 0 {
			cl = ncl
		}
	}

	for i, el := range els {
		var initwalkres *kchan
		var oldc *kchan

		for _, c := range cl {
			res, err := fWalk(c, mkFid(), []string{el})
			if err == nil {
				oldc = c
				initwalkres = res
				goto eval
			}
		}
		return nil, errors.New("name not found")

	eval:
		if el == ".." {
			// result is one possible result, but we need
			// to backwards-eval the mount table
			trimmedinitialnm := path.Dir(oldc.name)
			nmntres, err := p.mnt.reverseEval(initwalkres, trimmedinitialnm)
			if err == nil {
				cl = []*kchan{nmntres}
			} else {
				cl = []*kchan{initwalkres}
			}
		} else {
			if i == len(els)-1 && estop {
				cl = []*kchan{initwalkres}
			} else if ncl := p.mnt.forwardEval(initwalkres); len(ncl) > 0 {
				cl = ncl
			} else {
				cl = []*kchan{initwalkres}
			}
		}
	}

	return cl[0], nil
}
