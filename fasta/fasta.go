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

package fasta

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Read reads a FASTA file and returns the sequences matching the given filters.
func Read(r io.Reader, filters ...Filter) ([]Sequence, error) {
	var sequences []Sequence

	var description string
	var values []byte

	bufferedReader := bufio.NewReader(r)
	for {
		line, isPrefix, err := bufferedReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read fasta file: %w", err)
		}

		// Handle the case where the line is too long and is split across multiple ReadLine calls.
		for isPrefix {
			var nextLine []byte
			nextLine, isPrefix, err = bufferedReader.ReadLine()
			if err != nil {
				return nil, fmt.Errorf("failed to read continuation of line: %w", err)
			}
			line = append(line, nextLine...)
		}

		lineStr := string(line)
		if len(lineStr) == 0 {
			continue
		}

		if lineStr[0] == '>' {
			if len(description) > 0 {
				s := Sequence{
					Description: description,
					Values:      make([]byte, len(values)),
					index:       len(sequences),
				}

				copy(s.Values, values)

				if len(filters) > 0 {
					for _, filter := range filters {
						if filter(&s) {
							sequences = append(sequences, s)
							break
						}
					}
				} else {
					sequences = append(sequences, s)
				}
			}

			description = lineStr[1:]
			values = values[:0]
		} else {
			values = append(values, []byte(strings.ToUpper(lineStr))...)
		}
	}

	if len(values) > 0 {
		s := Sequence{
			Description: description,
			Values:      make([]byte, len(values)),
			index:       len(sequences),
		}

		copy(s.Values, values)

		if len(filters) > 0 {
			for _, filter := range filters {
				if filter(&s) {
					sequences = append(sequences, s)
					break
				}
			}
		} else {
			sequences = append(sequences, s)
		}
	}

	return sequences, nil
}

// Write writes the given sequences to a FASTA file.
func Write(w io.Writer, sequences []Sequence) error {
	for _, s := range sequences {
		_, err := fmt.Fprintf(w, ">%s\n", s.Description)
		if err != nil {
			return fmt.Errorf("failed to write fasta file: %w", err)
		}

		for i := 0; i < len(s.Values); i += 80 {
			end := min(i+80, len(s.Values))

			if _, err = w.Write(s.Values[i:end]); err != nil {
				return fmt.Errorf("failed to write fasta file: %w", err)
			}

			if end < len(s.Values) {
				if _, err = w.Write([]byte{'\n'}); err != nil {
					return fmt.Errorf("failed to write fasta file: %w", err)
				}
			}
		}

		if _, err = w.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("failed to write fasta file: %w", err)
		}
	}

	return nil
}

// Sequence represents a single sequence in a FASTA file.
type Sequence struct {
	Description string
	Values      []byte
	index       int
}

// Get returns the base at the given position.
func (s *Sequence) Get(i int64) (byte, error) {
	if i < 1 || i > int64(len(s.Values)) {
		return 0, fmt.Errorf("index out of range: %d", i)
	}

	return s.Values[i-1], nil
}

// GetRange returns the bases in the given position range.
func (s *Sequence) GetRange(start, end int64) ([]byte, error) {
	if start < 1 || start > int64(len(s.Values)) {
		return nil, fmt.Errorf("start index out of range: %d", start)
	}
	if end < 1 || end > int64(len(s.Values)) {
		return nil, fmt.Errorf("end index out of range: %d", end)
	}
	if start > end {
		return nil, fmt.Errorf("start index is greater than end index: %d > %d", start, end)
	}

	return s.Values[start-1 : end], nil
}

// Filter is a function that returns true if the given sequence should be included in the results.
type Filter func(*Sequence) bool

// FilterByID matches sequences with the given NCBI ID.
func FilterByID(id string) Filter {
	ncbiIDRegexp := regexp.MustCompile(`^([A-Z]{2}_[0-9]+\.[0-9]+)`)

	return func(s *Sequence) bool {
		match := ncbiIDRegexp.FindStringSubmatch(s.Description)

		return len(match) >= 1 && match[1] == id
	}
}

// FilterByIndex matches sequences with the given index.
func FilterByIndex(i int) Filter {
	return func(s *Sequence) bool {
		return s.index == i
	}
}
