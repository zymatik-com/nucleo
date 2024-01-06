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

package compress

import (
	"bytes"
	"compress/bzip2"
	"io"

	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	gzip "github.com/klauspost/pgzip"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

type autoDecompressingReadCloser struct {
	io.Reader
	close func() error
}

func Decompress(r io.Reader) (io.ReadCloser, error) {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}

	r = io.MultiReader(bytes.NewReader(buf[:n]), r)

	switch {
	case bytes.HasPrefix(buf, []byte{0x42, 0x5A, 0x68}): // BZIP2
		return &autoDecompressingReadCloser{
			Reader: bzip2.NewReader(r),
		}, nil
	case bytes.Equal(buf[0:2], []byte{0x1F, 0x8B}): // GZIP
		gzReader, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}

		return &autoDecompressingReadCloser{
			Reader: gzReader,
			close: func() error {
				if err := gzReader.Close(); err != nil {
					return err
				}

				return nil
			},
		}, nil
	case bytes.HasPrefix(buf, []byte{0x04, 0x22, 0x4D, 0x18}): // LZ4
		lz4Reader := lz4.NewReader(r)

		return &autoDecompressingReadCloser{
			Reader: lz4Reader,
		}, nil
	case bytes.HasPrefix(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}): // XZ
		xzReader, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}

		return &autoDecompressingReadCloser{
			Reader: xzReader,
		}, nil
	case bytes.HasPrefix(buf, []byte{0x78, 0x01}), bytes.HasPrefix(buf, []byte{0x78, 0x9C}), bytes.HasPrefix(buf, []byte{0x78, 0xDA}): // ZLIB
		zlibReader, err := zlib.NewReader(r)
		if err != nil {
			return nil, err
		}

		return &autoDecompressingReadCloser{
			Reader: zlibReader,
			close:  zlibReader.Close,
		}, nil
	case bytes.HasPrefix(buf, []byte{0x28, 0xB5, 0x2F, 0xFD}): // ZSTD
		zstdReader, err := zstd.NewReader(r)
		if err != nil {
			return nil, err
		}

		return &autoDecompressingReadCloser{
			Reader: zstdReader,
			close: func() error {
				zstdReader.Close()

				return nil
			},
		}, nil
	}

	return &autoDecompressingReadCloser{
		Reader: r,
	}, nil
}

func (r *autoDecompressingReadCloser) Close() error {
	if r.close != nil {
		return r.close()
	}

	return nil
}
