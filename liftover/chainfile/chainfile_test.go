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

package chainfile_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/brentp/vcfgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
func TestChainFile(t *testing.T) {
	ctx := context.Background()

	f, err := os.Open("../../testdata/GRCh37_to_GRCh38.chain.gz")
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

	grch37SNPs, err := readClinVarSNPs("../../testdata/clinvar_GRCh37_20231230.vcf.gz")
	require.NoError(t, err)

	grch38SNPs, err := readClinVarSNPs("../../testdata/clinvar_GRCh38_20231230.vcf.gz")
	require.NoError(t, err)

	var foundInBoth, successFullyLifted int
	for _, snp := range grch37SNPs {
		if _, ok := grch38SNPs[snp.id]; !ok {
			continue
		}

		foundInBoth++

		result, err := liftover.Lift(ctx, cf, "GRCh37", snp.chromosome, snp.position)
		if err != nil {
			continue
		}

		if result == grch38SNPs[snp.id].position {
			successFullyLifted++
		}
	}

	assert.Greater(t, successFullyLifted, 1000)
	assert.Greater(t, float64(successFullyLifted)/float64(foundInBoth), 0.995)
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
