/* SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * Zymatik Nucleo - A Bioinformatics library for Go.
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// Package chainfile provides utilities for reading chain files.
// Chain files are used to describe alignments between two reference genomes.
// https://genome.ucsc.edu/goldenPath/help/chain.html
package chainfile

import (
	"bufio"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"strconv"
	"strings"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/zymatik-com/genobase/types"
	"github.com/zymatik-com/nucleo/names"
)

// Chain represents a single Chain in a Chain file.
type Chain struct {
	Score       int64              // Alignment score.
	RefName     string             // Reference chromosome name.
	RefSize     int64              // Size of the reference chromosome.
	RefStrand   string             // Strand in the reference genome ('+' or '-').
	RefStart    int64              // Start position in the reference genome.
	RefEnd      int64              // End position in the reference genome.
	QueryName   string             // Query chromosome name.
	QuerySize   int64              // Size of the query chromosome.
	QueryStrand string             // Strand in the query genome ('+' or '-').
	QueryStart  int64              // Start position in the query genome.
	QueryEnd    int64              // End position in the query genome.
	ID_         int64              // Unique identifier for the chain.
	Alignments  augmentedtree.Tree // Interval tree of alignments.
}

func (c *Chain) LowAtDimension(dim uint64) int64 {
	return c.RefStart
}

func (c *Chain) HighAtDimension(dim uint64) int64 {
	return c.RefEnd
}

func (c *Chain) OverlapsAtDimension(with augmentedtree.Interval, dim uint64) bool {
	return true
}

func (c *Chain) ID() uint64 {
	return uint64(c.ID_)
}

// Alignment represents an Alignment block within a chain.
type Alignment struct {
	RefOffset   int64 `db:"ref_offset"`   // Offset of the aligned block in the reference chromosome from the start of the chain.
	QueryOffset int64 `db:"query_offset"` // Offset of the aligned block in the query chromosome from the start of the chain.
	Size        int64 `db:"size"`         // Size of the aligned block in bases.
}

func (a *Alignment) LowAtDimension(dim uint64) int64 {
	return a.RefOffset
}

func (a *Alignment) HighAtDimension(dim uint64) int64 {
	return a.RefOffset + a.Size
}

func (a *Alignment) OverlapsAtDimension(with augmentedtree.Interval, dim uint64) bool {
	return true
}

func (a *Alignment) ID() uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d;%d;%d", a.RefOffset, a.QueryOffset, a.Size)))
	return h.Sum64()
}

type Interval struct {
	Start int64
	End   int64
}

func (i *Interval) LowAtDimension(dim uint64) int64 {
	return i.Start
}

func (i *Interval) HighAtDimension(dim uint64) int64 {
	return i.End
}

func (i *Interval) OverlapsAtDimension(with augmentedtree.Interval, dim uint64) bool {
	return true
}

func (i *Interval) ID() uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d;%d", i.Start, i.End)))
	return h.Sum64()
}

// ChainFile represents a chain file.
type ChainFile struct {
	// ChainsByChromosome maps a chromosome name to an interval tree of chains.
	ChainsByChromosome map[string]augmentedtree.Tree
	// ChainByID maps a chain ID to a chain.
	ChainByID map[int64]*Chain
}

// Read loads a chain file from an io.Reader.
func Read(reader io.Reader) (*ChainFile, error) {
	chainFile := &ChainFile{
		ChainsByChromosome: make(map[string]augmentedtree.Tree),
		ChainByID:          make(map[int64]*Chain),
	}

	var currentChain *Chain
	var refOffset, queryOffset int64

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if strings.HasPrefix(line, "chain") {
			if currentChain != nil {
				tree, exists := chainFile.ChainsByChromosome[currentChain.RefName]
				if !exists {
					tree = augmentedtree.New(1)
				}
				tree.Add(currentChain)
				chainFile.ChainsByChromosome[currentChain.RefName] = tree
				chainFile.ChainByID[currentChain.ID_] = currentChain
			}

			if len(fields) < 12 {
				return nil, fmt.Errorf("invalid chain line: %s", line)
			}

			currentChain = &Chain{
				Score:       parseField(fields[1]),
				RefName:     names.Chromosome(fields[2]),
				RefSize:     parseField(fields[3]),
				RefStrand:   fields[4],
				RefStart:    parseField(fields[5]),
				RefEnd:      parseField(fields[6]),
				QueryName:   names.Chromosome(fields[7]),
				QuerySize:   parseField(fields[8]),
				QueryStrand: fields[9],
				QueryStart:  parseField(fields[10]),
				QueryEnd:    parseField(fields[11]),
				ID_:         parseField(fields[12]),
				Alignments:  augmentedtree.New(1),
			}

			// Reset the offsets.
			refOffset, queryOffset = 0, 0
		} else if currentChain != nil {
			// Parse an alignment block or the non-aligning region
			if len(fields) == 1 {
				size := parseField(fields[0])

				currentChain.Alignments.Add(&Alignment{
					RefOffset:   refOffset,
					QueryOffset: queryOffset,
					Size:        size,
				})

				refOffset += size
				queryOffset += size
			} else if len(fields) == 3 {
				size := parseField(fields[0])
				// Gap between this and the next block in the reference genome.
				refGap := parseField(fields[1])
				// Gap between this and the next block in the query genome.
				queryGap := parseField(fields[2])

				currentChain.Alignments.Add(&Alignment{
					RefOffset:   refOffset,
					QueryOffset: queryOffset,
					Size:        size,
				})

				refOffset += size + refGap
				queryOffset += size + queryGap
			} else {
				return nil, fmt.Errorf("invalid alignment line: %q", line)
			}
		}
	}

	if currentChain != nil {
		tree, exists := chainFile.ChainsByChromosome[currentChain.RefName]
		if !exists {
			tree = augmentedtree.New(1) // Create a new tree for this chromosome
		}
		tree.Add(currentChain)
		chainFile.ChainsByChromosome[currentChain.RefName] = tree
		chainFile.ChainByID[currentChain.ID_] = currentChain
	}

	return chainFile, scanner.Err()
}

// GetChain returns the chain for the given chromosome and position.
func (cf *ChainFile) GetChain(ctx context.Context, fromReference, chromosome string, position int64) (*types.Chain, error) {
	tree, ok := cf.ChainsByChromosome[chromosome]
	if !ok {
		return nil, fmt.Errorf("chromosome %s not found", chromosome)
	}

	query := &Interval{Start: position, End: position}
	intervals := tree.Query(query)
	if len(intervals) == 0 {
		return nil, fmt.Errorf("position %d not found in chromosome %s", position, chromosome)
	}

	chain := intervals[0].(*Chain)

	return &types.Chain{
		ID:          chain.ID_,
		Score:       chain.Score,
		RefName:     chain.RefName,
		RefSize:     chain.RefSize,
		RefStrand:   chain.RefStrand,
		RefStart:    chain.RefStart,
		RefEnd:      chain.RefEnd,
		QueryName:   chain.QueryName,
		QuerySize:   chain.QuerySize,
		QueryStrand: chain.QueryStrand,
		QueryStart:  chain.QueryStart,
		QueryEnd:    chain.QueryEnd,
	}, nil
}

// GetAlignment returns the alignment for the given chain and offset from the
// start of the chain.
func (cf *ChainFile) GetAlignment(ctx context.Context, chainID int64, offset int64) (*types.Alignment, error) {
	chain, ok := cf.ChainByID[chainID]
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	query := &Interval{Start: offset, End: offset}

	intervals := chain.Alignments.Query(query)
	if len(intervals) == 0 {
		return nil, fmt.Errorf("offset %d not found in chain %d", offset, chainID)
	}

	alignment := intervals[0].(*Alignment)

	return &types.Alignment{
		RefOffset:   alignment.RefOffset,
		QueryOffset: alignment.QueryOffset,
		Size:        alignment.Size,
	}, nil
}

func parseField(field string) int64 {
	value, err := strconv.ParseInt(field, 10, 64)
	if err != nil {
		return -1
	}

	return value
}
