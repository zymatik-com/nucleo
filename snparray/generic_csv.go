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

package snparray

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type genericCSVCodec struct{}

func (c *genericCSVCodec) Detect(r io.Reader) (bool, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false, scanner.Err()
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return strings.Contains(scanner.Text(), ",") &&
		!strings.Contains(scanner.Text(), "\t"), nil
}

type genericCSVReader struct {
	reader         *csv.Reader
	columnMappings map[string]int
}

func (c *genericCSVCodec) Open(r io.Reader) (Reader, error) {
	reader := csv.NewReader(r)
	reader.Comment = '#'

	record, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading genome file: %w", err)
	}

	// TODO: guess column mappings if not present.

	columnMappings := make(map[string]int)
	for i, colName := range record {
		columnMappings[strings.ToLower(strings.TrimSpace(colName))] = i
	}

	return &genericCSVReader{
		reader:         reader,
		columnMappings: columnMappings,
	}, nil
}

func (r *genericCSVReader) Read() (*SNP, error) {
	var record []string

	// Skip over no call variants.
	genotype := "--"
	for genotype == "--" || genotype == "00" {
		var err error
		record, err = r.reader.Read()
		if err != nil {
			return nil, err
		}

		if len(record) < len(r.columnMappings) {
			return nil, fmt.Errorf("not enough columns")
		}

		genotype = record[r.columnMappings["result"]]
	}

	// TODO: support a more fuzzy matching of column names.

	position, err := strconv.ParseInt(record[r.columnMappings["position"]], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing position: %s", err)
	}

	return &SNP{
		RSID:       record[r.columnMappings["rsid"]],
		Chromosome: record[r.columnMappings["chromosome"]],
		Position:   position,
		Genotype:   genotype,
	}, nil
}
