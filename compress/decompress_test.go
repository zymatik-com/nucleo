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
		"testdata/test.bgz",
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
