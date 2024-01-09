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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/zymatik-com/genobase/types"
	"github.com/zymatik-com/nucleo/names"
)

type twentyThreeAndMeCodec struct{}

func (c *twentyThreeAndMeCodec) Detect(r io.Reader) (bool, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false, scanner.Err()
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return strings.Contains(scanner.Text(), "23andMe"), nil
}

type twentyThreeAndMeReader struct {
	reader         *csv.Reader
	columnMappings map[string]int
}

func (c *twentyThreeAndMeCodec) Open(r io.Reader) (Reader, error) {
	var buf bytes.Buffer
	var lastCommentLine string

	bufReader := bufio.NewReader(r)
	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error reading genome file: %w", err)
			}
			if err == io.EOF && line == "" {
				break
			}
		}

		buf.WriteString(line)
		buf.WriteString("\n")

		if !strings.HasPrefix(line, "#") {
			break
		}

		lastCommentLine = line
	}

	if lastCommentLine == "" {
		return nil, fmt.Errorf("header comment not found")
	}

	columnMappings := make(map[string]int)
	for i, colName := range strings.Split(strings.TrimPrefix(lastCommentLine, "#"), "\t") {
		columnMappings[strings.ToLower(strings.TrimSpace(colName))] = i
	}

	csvReader := csv.NewReader(io.MultiReader(&buf, bufReader))
	csvReader.Comma = '\t'
	csvReader.Comment = '#'

	return &twentyThreeAndMeReader{
		reader:         csvReader,
		columnMappings: columnMappings,
	}, nil
}

func (r *twentyThreeAndMeReader) Reference() types.Reference {
	return types.ReferenceGRCh37
}

func (r *twentyThreeAndMeReader) Read() (*SNP, error) {
	var record []string

	// Skip over no call variants.
	genotype := "--"
	for genotype == "--" {
		var err error
		record, err = r.reader.Read()
		if err != nil {
			return nil, err
		}

		if len(record) < len(r.columnMappings) {
			return nil, fmt.Errorf("not enough columns")
		}

		genotype = record[r.columnMappings["genotype"]]
	}

	position, err := strconv.ParseInt(record[r.columnMappings["position"]], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing position: %s", err)
	}

	return &SNP{
		RSID:       record[r.columnMappings["rsid"]],
		Chromosome: names.Chromosome(record[r.columnMappings["chromosome"]]),
		Position:   position,
		Genotype:   genotype,
	}, nil
}
