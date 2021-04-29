package client

import (
	"errors"
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
}

// Unbind, returning an error if no such mapping exists
func (m *mountTable) unbind(from *kchan, to *kchan) error {
	for i, mp := range m.tbl {
		if kchanCmp(mp.from, from) {
			if kchanCmp(mp.to, to) {
				// from->to exists
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
			results = append(results, mp.to)
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
				return mp.from, nil
			}
		}
	}

	return &kchan{}, errors.New("failed to find match")
}
