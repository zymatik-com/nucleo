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

package compress

import (
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

type autoCompressingWriteCloser struct {
	io.WriteCloser
}

// Guess the compression algorithm based on the file extension.
// If none is found, use gzip.
func Compress(name string, w io.Writer) (io.WriteCloser, error) {
	switch {
	case strings.HasSuffix(name, ".lz4"):
		lz4Writer := lz4.NewWriter(w)

		return &autoCompressingWriteCloser{
			WriteCloser: lz4Writer,
		}, nil
	case strings.HasSuffix(name, ".xz"):
		xzWriter, err := xz.NewWriter(w)
		if err != nil {
			return nil, err
		}

		return &autoCompressingWriteCloser{
			WriteCloser: xzWriter,
		}, nil
	case strings.HasSuffix(name, ".zst"):
		zstdWriter, err := zstd.NewWriter(w)
		if err != nil {
			return nil, err
		}

		return &autoCompressingWriteCloser{
			WriteCloser: zstdWriter,
		}, nil
	default:
		gzWriter := gzip.NewWriter(w)

		return &autoCompressingWriteCloser{
			WriteCloser: gzWriter,
		}, nil
	}
}
