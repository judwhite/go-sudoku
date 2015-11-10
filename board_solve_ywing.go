package main

import (
	"fmt"
	"sync"

	"github.com/judwhite/go-sudoku/internal/bits"
)

func (b *board) SolveYWing() error {
	// http://www.sudokuwiki.org/Y_Wing_Strategy
	// start simple..
	// look for three cells which each have two hints and can 'see' each other,
	// like 3 corners of an xwing.
	// the three cells have hints AB,BC,CB.. any C in the 4th corner can be removed.

	// define a 'hinge' cell and the two 'wing' cells
	// the wings can be 'seen' by the hinge
	// the value NOT in the hinge ('C' for example) can be taken off any
	// cells which can be seen by both wing cells

	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}

		blit := b.blits[i]
		if bits.GetNumberOfSetBits(blit) != 2 {
			continue
		}

		visibleToHinge := b.getVisibleCells(i)

		// filter visible list to only those which have two set bits
		// and have ONLY one in common with the hinge
		var candidates []int
		for _, item := range visibleToHinge {
			itemBlit := b.blits[item]
			if bits.GetNumberOfSetBits(itemBlit) != 2 {
				continue
			}
			if !bits.HasSingleBit(blit & itemBlit) {
				continue
			}
			candidates = append(candidates, item)
		}

		// get all permutations of the candidates
		perms := getPermutations(2, candidates, []int{})

		// filter permutations where the candidates share only one hint
		var wingsList [][]int
		for _, list := range perms {
			if len(list) != 2 {
				fmt.Println("len(list) != 2 ???")
				continue
			}
			wingBlit1 := b.blits[list[0]]
			wingBlit2 := b.blits[list[1]]

			if bits.HasSingleBit(wingBlit1&wingBlit2) &&
				bits.GetNumberOfSetBits(blit|wingBlit1|wingBlit2) == 3 {
				wingsList = append(wingsList, list)
			}
		}

		if len(wingsList) == 0 {
			continue
		}

		for idx, wings := range wingsList {
			if len(wings) != 2 {
				// TODO: len(wings) should always be 2, being defensive
				continue
			}

			sum := b.blits[wings[0]] | b.blits[wings[1]]
			targets := b.getVisibleCells(wings[0])
			targets = intersect(targets, b.getVisibleCells(wings[1]))

			if len(targets) == 0 {
				continue
			}

			removeHint := sum & ^blit

			var once1 sync.Once
			print1 := func() {
				fmt.Printf("* %#2v %s\n", getCoords(i), bits.GetString(blit))
				fmt.Printf("wing set %d:\n", idx+1)
				fmt.Printf("-- %#2v %s\n", getCoords(wings[0]), bits.GetString(b.blits[wings[0]]))
				fmt.Printf("-- %#2v %s\n", getCoords(wings[1]), bits.GetString(b.blits[wings[1]]))
				fmt.Printf("-- remove hint: %d\n", bits.GetSingleBitValue(removeHint))
				fmt.Printf("-- targets:\n")
			}

			updated := false
			for _, target := range targets {
				if target == i || target == wings[0] || target == wings[1] {
					continue
				}
				once1.Do(print1)
				fmt.Printf("---- %#2v %s\n", getCoords(target), bits.GetString(b.blits[target]))
				if b.blits[target]&removeHint == removeHint {
					updated = true
					if err := b.updateCandidates(target, i, ^removeHint); err != nil {
						return err
					}
				}
			}
			if updated {
				// let simpler techniques take over
				return nil
			}
		}
	}

	return nil
}

func (b *board) getVisibleCells(pos int) []int {
	var list []int
	coords := getCoords(pos)

	for i := 0; i < 81; i++ {
		if i == pos || b.solved[i] != 0 {
			continue
		}
		t := getCoords(i)
		if t.row == coords.row ||
			t.col == coords.col ||
			t.box == coords.box {
			list = append(list, i)
		}
	}

	return list
}

func (b *board) getVisibleCellsWithHint(pos int, hint uint) []int {
	var list []int
	coords := getCoords(pos)

	for i := 0; i < 81; i++ {
		if i == pos || b.solved[i] != 0 || b.blits[i]&hint != hint {
			continue
		}
		t := getCoords(i)
		if t.row == coords.row ||
			t.col == coords.col ||
			t.box == coords.box {
			list = append(list, i)
		}
	}

	return list
}
