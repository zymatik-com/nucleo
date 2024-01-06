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

package compress_test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zymatik-com/nucleo/compress"
)

func TestAutoDecompressingReadCloser(t *testing.T) {
	paths := []string{
		"testdata/test.bz2",
		"testdata/test.gz",
		"testdata/test.lz4",
		"testdata/test.xz",
		"testdata/test.zlib",
		"testdata/test.zst",
		"testdata/test.txt",
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			f, err := os.Open(p)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, f.Close())
			})

			dr, err := compress.Decompress(f)
			require.NoError(t, err)

			buf, err := io.ReadAll(dr)
			require.NoError(t, err)

			assert.Equal(t, "Hello, World!\n", string(buf))

			require.NoError(t, dr.Close())
		})
	}
}
