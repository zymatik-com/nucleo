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

// Package snparray provides readers for common SNP array file formats.
// Such as direct-to-consumer genetic testing services like 23andMe and AncestryDNA.
package snparray

import (
	"bytes"
	"fmt"
	"io"

	"github.com/zymatik-com/genobase/types"
)

type SNP struct {
	RSID       string
	Chromosome string
	Position   int64
	Genotype   string
}

// Codec is a SNP array file format encoder/decoder.
type Codec interface {
	// Detect returns true if the file format is detected.
	Detect(r io.Reader) (bool, error)
	// Open opens the SNP array file and returns a lazy SNP reader.
	Open(r io.Reader) (Reader, error)
}

// Reader is a lazy SNP reader.
type Reader interface {
	// Reference returns the reference assembly used by the SNP array.
	Reference() types.Reference
	// Read reads the next SNP from the file. It returns io.EOF if there are no
	// more SNPs.
	Read() (*SNP, error)
}

var codecs = []Codec{
	&twentyThreeAndMeCodec{},
	&ancestryDNACodec{},
	&genericCSVCodec{},
	&genericTSVCodec{},
}

// Open opens the SNP array file and returns a lazy SNP reader.
func Open(r io.Reader) (Reader, error) {
	// Peak at the first few lines to determine the file format.
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}

	for _, codec := range codecs {
		ok, err := codec.Detect(bytes.NewReader(buf[:n]))
		if err != nil {
			return nil, err
		}

		if ok {
			return codec.Open(io.MultiReader(bytes.NewReader(buf[:n]), r))
		}
	}

	return nil, fmt.Errorf("unknown snparray format")
}
