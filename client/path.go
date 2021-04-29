package client

import (
	"errors"
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
func (p *proc) evaluate(pth string, estop bool) (*kchan, error) {
	els := strings.Split(cleanPath(pth), "/")
	cl := []*kchan{&rootChannel}

	if pth == "." {
		return p.cwd, nil
	} else if pth == "/" {
		return &rootChannel, nil
	} else if pth[0] != '/' {
		cl = []*kchan{p.cwd}
	}

	for i, el := range els {
		var initwalkres *kchan
		var oc *kchan

		for _, c := range cl {
			res, err := fWalk(c, mkFid(), []string{el})
			if err == nil {
				oc = c
				initwalkres = res
				goto eval
			}
		}
		return nil, errors.New("name not found")

	eval:
		if el == ".." {
			// result is one possible result, but we need
			// to backwards-eval the mount table
			trimmedinitialnm := path.Dir(oc.name)
			nmntres, err := p.mnt.reverseEval(initwalkres, trimmedinitialnm)
			if err == nil {
				cl = []*kchan{nmntres}
			} else {
				cl = []*kchan{initwalkres}
			}
		} else {
			if i == len(els)-1 && estop {
				cl = []*kchan{initwalkres}
			} else if ncl := p.mnt.forwardEval(initwalkres); ncl != nil {
				cl = ncl
			} else {
				cl = []*kchan{initwalkres}
			}
		}
	}

	return cl[0], nil
}
