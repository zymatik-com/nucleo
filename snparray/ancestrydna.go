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

package snparray

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/zymatik-com/genobase/types"
	"github.com/zymatik-com/nucleo/names"
)

type ancestryDNACodec struct{}

func (c *ancestryDNACodec) Detect(r io.Reader) (bool, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false, scanner.Err()
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return strings.Contains(scanner.Text(), "AncestryDNA"), nil
}

type ancestryDNAReader struct {
	reader         *csv.Reader
	columnMappings map[string]int
}

func (c *ancestryDNACodec) Open(r io.Reader) (Reader, error) {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.Comment = '#'

	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading genome file: %w", err)
	}

	columnMappings := make(map[string]int)
	for i, colName := range record {
		columnMappings[strings.ToLower(strings.TrimSpace(colName))] = i
	}

	return &ancestryDNAReader{
		reader:         reader,
		columnMappings: columnMappings,
	}, nil
}

func (r *ancestryDNAReader) Reference() types.Reference {
	return types.ReferenceGRCh37
}

func (r *ancestryDNAReader) Read() (*SNP, error) {
	var record []string

	// Skip over no call variants.
	genotype := "00"
	for genotype == "00" {
		var err error
		record, err = r.reader.Read()
		if err != nil {
			return nil, err
		}

		if len(record) < len(r.columnMappings) {
			return nil, fmt.Errorf("not enough columns")
		}

		genotype = record[r.columnMappings["allele1"]] + record[r.columnMappings["allele2"]]
	}

	position, err := strconv.ParseInt(record[r.columnMappings["position"]], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing position: %s", err)
	}

	chromosome := names.Chromosome(record[r.columnMappings["chromosome"]])

	// AncestryDNA has a very interesting chromosome naming convention.
	if chromosome == "23" {
		chromosome = "X"
	} else if chromosome == "24" {
		chromosome = "Y"
	} else if chromosome == "25" {
		chromosome = "PAR"

		switch r.Reference() {
		case types.ReferenceGRCh37:
			if position >= 154931044 {
				chromosome = "PAR2"
			}
		case types.ReferenceGRCh38:
			if position >= 155701383 {
				chromosome = "PAR2"
			}
		}
	} else if chromosome == "26" {
		chromosome = "MT"
	}

	return &SNP{
		RSID:       record[r.columnMappings["rsid"]],
		Chromosome: chromosome,
		Position:   position,
		Genotype:   genotype,
	}, nil
}
