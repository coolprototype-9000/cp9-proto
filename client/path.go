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
	if els[0] == "" {
		els = els[1:]
	}
	cl := []*kchan{&rootChannel}
	var canClunk bool

	if pth == "." {
		if estop {
			return p.cwd, nil
		} else {
			cl = p.mnt.forwardEval(p.cwd)
			if len(cl) > 0 {
				goto done
			}
			return p.cwd, nil
		}
	} else if pth == "/" {
		if estop {
			return &rootChannel, nil
		} else {
			cl = p.mnt.forwardEval(&rootChannel)
			if len(cl) > 0 {
				goto done
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
		var cc *kchan

		for _, c := range cl {
			res, err := fWalk(c, mkFid(), []string{el})
			if err == nil {
				if canClunk {
					fClunk(c)
				}
				cc = c
				initwalkres = res
				goto eval
			}
		}
		return nil, errors.New("name not found")

	eval:
		if el == ".." {
			// result is one possible result, but we need
			// to backwards-eval the mount table
			trimmedinitialnm := path.Dir(cc.name)
			nmntres, err := p.mnt.reverseEval(initwalkres, trimmedinitialnm)
			if err == nil {
				cl = []*kchan{nmntres}
				canClunk = false
			} else {
				cl = []*kchan{initwalkres}
				canClunk = true
			}
		} else {
			fmt.Printf("Evaluating: %v\n", *initwalkres)
			if i == len(els)-1 && estop {
				cl = []*kchan{initwalkres}
				canClunk = true
			} else if ncl := p.mnt.forwardEval(initwalkres); len(ncl) > 0 {
				cl = ncl
				canClunk = false
			} else {
				cl = []*kchan{initwalkres}
				canClunk = true
			}
		}
	}
done:
	if !canClunk {
		// Dup the fid if it's a permanent ref in
		// the mount table, so that the user can safely
		// get rid of it
		nn, err := fWalk(cl[0], mkFid(), []string{})
		if err != nil {
			return nil, errors.New("failed to dup fd")
		}
		cl[0] = nn
	}
	return cl[0], nil
}
