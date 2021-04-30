package client

import (
	"errors"
	"fmt"
)

type mountPair struct {
	from *kchan
	to   *kchan
}

type mountTable struct {
	tbl []mountPair
}

func mkFreshMountTable() *mountTable {
	mt := &mountTable{
		tbl: []mountPair{},
	}
	return mt
}

// Add an entry to the mount table. Succeeds no matter what
func (m *mountTable) bind(from *kchan, to *kchan, first bool) {
	np := mountPair{from, to}
	if !first {
		m.tbl = append(m.tbl, np)
	} else {
		m.tbl = append([]mountPair{np}, m.tbl...)
	}
	fmt.Printf("!: %v -> %v\n", *from, *to)
}

// Unbind, returning an error if no such mapping exists
func (m *mountTable) unbind(from *kchan, to *kchan) error {
	for i, mp := range m.tbl {
		if kchanCmp(mp.from, from) {
			if kchanCmp(mp.to, to) {
				// from->to exists
				fmt.Printf("*: %v -> %v\n", *mp.from, *mp.to)
				m.tbl = append(m.tbl[:i], m.tbl[i+1:]...)
				return nil
			}
		}
	}

	return errors.New("no such mapping")
}

// Given a starting directory, return in order all
// corresponding to entries
func (m *mountTable) forwardEval(from *kchan) []*kchan {
	results := []*kchan{}
	for _, mp := range m.tbl {
		if kchanCmp(mp.from, from) {
			var nc *kchan = new(kchan)
			*nc = *mp.to
			results = append(results, nc)
		}
	}
	return results
}

// Given a to entry, and a lexical name of a from directory
// to search for, return a match if it exists
func (m *mountTable) reverseEval(to *kchan, from string) (*kchan, error) {
	for _, mp := range m.tbl {
		if kchanCmp(mp.to, to) {
			if mp.from.name == from {
				fmt.Printf("%v -> %v\n", *mp.from, *mp.to)
				var nc *kchan = new(kchan)
				*nc = *mp.from
				return nc, nil
			}
		}
	}

	return &kchan{}, errors.New("failed to find match")
}
