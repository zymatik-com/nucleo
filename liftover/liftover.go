/* SPDX-License-Identifier: MPL-2.0
 *
 * Zymatik Nucleo - A Bioinformatics library for Go.
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the Mozilla Public License v2.0.
 *
 * You should have received a copy of the Mozilla Public License v2.0
 * along with this program. If not, see <https://mozilla.org/MPL/2.0/>.
 */

// Package liftover provides utilities for lifting genomic coordinates from
// one reference genome to another, a process known as "lifting over".
// It leverages a database to store and retrieve the necessary chain and
// alignment information.
package liftover

import (
	"context"
	"fmt"

	"github.com/Workiva/go-datastructures/augmentedtree"
	"github.com/cheggaaa/pb/v3"
	"github.com/zymatik-com/genobase"
	"github.com/zymatik-com/genobase/types"
	"github.com/zymatik-com/nucleo/liftover/chainfile"
)

// ChainSource is a source of chain and alignment information.
type ChainSource interface {
	// GetChain returns the chain for the given chromosome and position.
	GetChain(ctx context.Context, from types.Reference, chromosome types.Chromosome, position int64) (*types.Chain, error)
	// GetAlignment returns the alignment for the given chain and offset from the
	// start of the chain.
	GetAlignment(ctx context.Context, chainID int64, offset int64) (*types.Alignment, error)
}

// Lift returns the position in the query genome for the given position in the
// reference genome.
func Lift(ctx context.Context, src ChainSource, from types.Reference, chromosome types.Chromosome, position int64) (int64, error) {
	chain, err := src.GetChain(ctx, from, chromosome, position)
	if err != nil {
		return -1, fmt.Errorf("could not get chain: %w", err)
	}

	alignment, err := src.GetAlignment(ctx, chain.ID, position-chain.RefStart)
	if err != nil {
		return -1, fmt.Errorf("position %d not found in chromosome %s: %w", position, chromosome, err)
	}

	queryPosition := chain.QueryStart + alignment.QueryOffset
	if chain.QueryStrand == "+" {
		queryPosition += position - (chain.RefStart + alignment.RefOffset)
	} else {
		queryPosition = chain.QueryEnd - (chain.RefStart + alignment.RefOffset)
	}

	return queryPosition, nil
}

// StoreChainFile stores the chain file in the database in a queryable format.
func StoreChainFile(ctx context.Context, db *genobase.DB, from types.Reference, cf *chainfile.ChainFile, showProgress bool) error {
	var bar *pb.ProgressBar
	if showProgress {
		total := 0
		for _, chains := range cf.ChainsByChromosome {
			total += int(chains.Len())
		}

		bar = pb.StartNew(total)
		defer bar.Finish()
	}

	for _, chains := range cf.ChainsByChromosome {
		var storeErr error
		chains.Traverse(func(interval augmentedtree.Interval) {
			if storeErr != nil {
				return
			}

			if bar != nil {
				bar.Increment()
			}

			chain := interval.(*chainfile.Chain)

			chainID, err := db.StoreChain(ctx, from, &types.Chain{
				Score:       chain.Score,
				Ref:         from,
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
			})
			if err != nil {
				storeErr = fmt.Errorf("could not store chain: %w", err)
				return
			}

			dbAlignments := make([]types.Alignment, chain.Alignments.Len())

			chain.Alignments.Traverse(func(interval augmentedtree.Interval) {
				alignment := interval.(*chainfile.Alignment)

				dbAlignments = append(dbAlignments, types.Alignment{
					RefOffset:   alignment.RefOffset,
					QueryOffset: alignment.QueryOffset,
					Size:        alignment.Size,
				})
			})

			if err := db.StoreAlignments(ctx, chainID, dbAlignments); err != nil {
				storeErr = fmt.Errorf("could not store alignments: %w", err)
				return
			}
		})
		if storeErr != nil {
			return storeErr
		}
	}

	return nil
}
