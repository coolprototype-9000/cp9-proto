package client

import (
	"path"
)

// This legitimately brought a smile to my face
// Thanks, plan 9 / golang crossover
func cleanPath(orig string) string {
	return path.Clean(orig)
}

// Walk a fresh fid to the path in question,
// returning a fresh kchan. This kchan should
// be garbage collected if it is not active
// post evaluation. The path does NOT have
// to be clean, we do that just in case.
/*
func (p *proc) evaluate(path string) kchan {
	cp := cleanPath(path)
	if cp == "." {
		return p.cwd
	} else if cp == "/" {
		return rootChannel
	}

	var c kchan
	if cp[0] == '/' {
		c = rootChannel
	} else {
		c = p.cwd
	}

	steps := strings.Split(cp, "/")
	for _, step := range steps {

	}
}

*/
