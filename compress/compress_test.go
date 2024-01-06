/* SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * Zymatik Nucleo - A bioinformatics library for Go (focused on Human Genomics).
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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zymatik-com/nucleo/compress"
)

func TestAutoCompressingWriteCloser(t *testing.T) {
	names := []string{
		"test.gz",
		"test.lz4",
		"test.xz",
		"test.zst",
	}

	dir := t.TempDir()
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)

			f, err := os.Create(path)
			require.NoError(t, err)

			w, err := compress.Compress(name, f)
			require.NoError(t, err)

			_, err = w.Write([]byte("Hello, World!\n"))
			require.NoError(t, err)

			require.NoError(t, w.Close())

			require.NoError(t, f.Close())

			f, err = os.Open(path)
			require.NoError(t, err)

			dr, err := compress.Decompress(f)
			require.NoError(t, err)

			buf, err := io.ReadAll(dr)
			require.NoError(t, err)

			require.NoError(t, dr.Close())

			assert.Equal(t, "Hello, World!\n", string(buf))
		})
	}
}
