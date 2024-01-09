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

package liftover_test

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/brentp/vcfgo"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zymatik-com/genobase"
	"github.com/zymatik-com/genobase/types"
	"github.com/zymatik-com/nucleo/compress"
	"github.com/zymatik-com/nucleo/liftover"
	"github.com/zymatik-com/nucleo/liftover/chainfile"
)

// A simple test to check if the liftover works as expected by validating the
// results against the ClinVar database (which is published for GRCh37 and GRCh38).
// Inspired by the approach described in [1].
//
// References:
//  1. Park, K.-J.; Yoon, Y.A.; Park, J.-H. Evaluation of Liftover Tools for the
//     Conversion of Genome Reference Consortium Human Build 37 to Build 38 Using
//     ClinVar Variants. Genes 2023, 14, 1875. https://doi.org/10.3390/genes14101875.
func TestLiftOver(t *testing.T) {
	ctx := context.Background()
	db, err := genobase.Open(ctx, slogt.New(t), "")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	// Initialize database from a chain file.
	{
		f, err := os.Open("../testdata/GRCh37_to_GRCh38.chain.gz")
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		dr, err := compress.Decompress(f)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, dr.Close())
		})

		cf, err := chainfile.Read(dr)
		require.NoError(t, err)

		err = liftover.StoreChainFile(ctx, db, types.ReferenceGRCh37, cf, false)
		require.NoError(t, err)
	}

	{
		f, err := os.Open("../testdata/NCBI36_to_GRCh38.chain.gz")
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, f.Close())
		})

		dr, err := compress.Decompress(f)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, dr.Close())
		})

		cf, err := chainfile.Read(dr)
		require.NoError(t, err)

		err = liftover.StoreChainFile(ctx, db, types.ReferenceNCBI36, cf, false)
		require.NoError(t, err)
	}

	t.Run("NCBI36 To GRCh38", func(t *testing.T) {
		ncbi36SNPs, err := readLegacySNPs("../testdata/snp130.txt.gz")
		require.NoError(t, err)

		grch38SNPs, err := readClinVarSNPs("../testdata/clinvar_GRCh38_20231230.vcf.gz")
		require.NoError(t, err)

		var foundInBoth, successFullyLifted int
		for _, snp := range ncbi36SNPs {
			if _, ok := grch38SNPs[snp.id]; !ok {
				continue
			}

			foundInBoth++

			result, err := liftover.Lift(ctx, db, types.ReferenceNCBI36, snp.chromosome, snp.position)
			if err != nil {
				continue
			}

			if result == grch38SNPs[snp.id].position {
				successFullyLifted++
			}
		}

		// Lifting this ancient dbSNP build from 2009 to GRCh38 is not perfect.
		// Upon inspection, I've checked the majority against other liftover tools
		// and we seem consistent.
		assert.Greater(t, successFullyLifted, 500)
		assert.Greater(t, float64(successFullyLifted)/float64(foundInBoth), 0.85)
	})

	t.Run("GRCh37 To GRCh38", func(t *testing.T) {
		grch37SNPs, err := readClinVarSNPs("../testdata/clinvar_GRCh37_20231230.vcf.gz")
		require.NoError(t, err)

		grch38SNPs, err := readClinVarSNPs("../testdata/clinvar_GRCh38_20231230.vcf.gz")
		require.NoError(t, err)

		var foundInBoth, successFullyLifted int
		for _, snp := range grch37SNPs {
			if _, ok := grch38SNPs[snp.id]; !ok {
				continue
			}

			foundInBoth++

			result, err := liftover.Lift(ctx, db, types.ReferenceGRCh37, snp.chromosome, snp.position)
			if err != nil {
				continue
			}

			if result == grch38SNPs[snp.id].position {
				successFullyLifted++
			}
		}

		assert.Greater(t, successFullyLifted, 1000)
		assert.Greater(t, float64(successFullyLifted)/float64(foundInBoth), 0.995)
	})
}

type snp struct {
	id         int64
	chromosome string
	position   int64
}

func readClinVarSNPs(path string) (map[int64]snp, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dr, err := compress.Decompress(f)
	if err != nil {
		return nil, err
	}
	defer dr.Close()

	vcfReader, err := vcfgo.NewReader(dr, false)
	if err != nil {
		return nil, err
	}

	snps := make(map[int64]snp)
	for {
		variant := vcfReader.Read()
		if variant == nil {
			break
		}

		idStr, err := variant.Info().Get("RS")
		if err != nil {
			continue
		}

		id, err := strconv.ParseInt(idStr.(string), 10, 64)
		if err != nil {
			continue
		}

		snps[id] = snp{
			id:         id,
			chromosome: strings.ToUpper(strings.TrimPrefix(variant.Chromosome, "chr")),
			position:   int64(variant.Pos),
		}
	}

	return snps, nil
}

func readLegacySNPs(path string) (map[int64]snp, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dr, err := compress.Decompress(f)
	if err != nil {
		return nil, err
	}
	defer dr.Close()

	tsvReader := csv.NewReader(dr)
	tsvReader.Comma = '\t'

	snps := make(map[int64]snp)
	for {
		record, err := tsvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		id, err := strconv.ParseInt(strings.TrimPrefix(record[4], "rs"), 10, 64)
		if err != nil {
			continue
		}

		position, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			continue
		}

		chromosome := strings.ToUpper(strings.TrimPrefix(record[1], "chr"))
		if chromosome == "M" {
			chromosome = "MT"
		}

		snps[id] = snp{
			id:         id,
			chromosome: chromosome,
			// The position is 0-based in the legacy file.
			position: position + 1,
		}
	}

	return snps, nil
}
