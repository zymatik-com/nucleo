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
