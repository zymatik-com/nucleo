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

package snparray_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zymatik-com/nucleo/compress"
	"github.com/zymatik-com/nucleo/snparray"
)

func TestSNPArrayOpen(t *testing.T) {
	t.Run("23andMe", func(t *testing.T) {
		f, err := os.Open("../testdata/huBE0518_23andMe.txt.gz")
		require.NoError(t, err)
		t.Cleanup(func() { f.Close() })

		dr, err := compress.Decompress(f)
		require.NoError(t, err)
		t.Cleanup(func() { dr.Close() })

		snpReader, err := snparray.Open(dr)
		require.NoError(t, err)

		snp, err := snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs548049170", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(69869), snp.Position)
		assert.Equal(t, "TT", snp.Genotype)

		snp, err = snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs9283150", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(565508), snp.Position)
		assert.Equal(t, "AA", snp.Genotype)
	})

	t.Run("Ancestry DNA", func(t *testing.T) {
		f, err := os.Open("../testdata/huBE0518_AncestryDNA.txt.gz")
		require.NoError(t, err)
		t.Cleanup(func() { f.Close() })

		dr, err := compress.Decompress(f)
		require.NoError(t, err)
		t.Cleanup(func() { dr.Close() })

		snpReader, err := snparray.Open(dr)
		require.NoError(t, err)

		snp, err := snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs3131972", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(752721), snp.Position)
		assert.Equal(t, "AA", snp.Genotype)

		snp, err = snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs114525117", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(759036), snp.Position)
		assert.Equal(t, "GG", snp.Genotype)
	})

	t.Run("MyHeritage", func(t *testing.T) {
		f, err := os.Open("../testdata/hu545C8F_ftdna.csv.gz")
		require.NoError(t, err)
		t.Cleanup(func() { f.Close() })

		dr, err := compress.Decompress(f)
		require.NoError(t, err)
		t.Cleanup(func() { dr.Close() })

		snpReader, err := snparray.Open(dr)
		require.NoError(t, err)

		snp, err := snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs4477212", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(72017), snp.Position)
		assert.Equal(t, "AA", snp.Genotype)

		snp, err = snpReader.Read()
		require.NoError(t, err)

		assert.Equal(t, "rs3131972", snp.RSID)
		assert.Equal(t, "1", snp.Chromosome)
		assert.Equal(t, int64(742584), snp.Position)
		assert.Equal(t, "GG", snp.Genotype)
	})
}