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

package fasta_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zymatik-com/nucleo/compress"
	"github.com/zymatik-com/nucleo/fasta"
)

func TestFastARead(t *testing.T) {
	f, err := os.Open("testdata/GCF_000182965.3_ASM18296v3_genomic.fna.gz")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	dr, err := compress.Decompress(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dr.Close())
	})

	sequences, err := fasta.Read(dr, fasta.FilterByID("NC_032093.1"))
	require.NoError(t, err)

	assert.Len(t, sequences, 1)

	s := sequences[0]
	assert.Equal(t, "NC_032093.1 Candida albicans SC5314 chromosome 5, complete sequence", s.Description)

	bases, err := s.GetRange(1, 10)
	require.NoError(t, err)

	assert.Equal(t, []byte("CAACCTATAT"), bases)

	base, err := s.Get(11)
	require.NoError(t, err)

	assert.Equal(t, byte('A'), base)
}

func TestFastAWrite(t *testing.T) {
	f, err := os.Open("testdata/GCF_000182965.3_ASM18296v3_genomic.fna.gz")
	require.NoError(t, err)

	dr, err := compress.Decompress(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dr.Close())
	})

	expectedSequences, err := fasta.Read(dr)
	require.NoError(t, err)

	outPath := filepath.Join(t.TempDir(), "test.fna.gz")
	f, err = os.Create(outPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	cr, err := compress.Compress(filepath.Base(outPath), f)
	require.NoError(t, err)

	require.NoError(t, fasta.Write(cr, expectedSequences))
	require.NoError(t, cr.Close())

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err)

	dr, err = compress.Decompress(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dr.Close())
	})

	sequences, err := fasta.Read(dr)
	require.NoError(t, err)

	assert.Len(t, sequences, len(expectedSequences))

	for i, s := range expectedSequences {
		assert.Equal(t, s.Description, sequences[i].Description)
		assert.Equal(t, s.Values, sequences[i].Values)
	}
}
