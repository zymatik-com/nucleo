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
