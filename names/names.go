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

package names

import (
	"fmt"
	"strings"

	"github.com/zymatik-com/genobase/types"
)

// Chromosome returns a sanitized/standardized chromosome name.
func Chromosome(chromosome string) string {
	chromosome = strings.ToUpper(strings.TrimPrefix(chromosome, "chr"))
	if chromosome == "M" {
		chromosome = "MT"
	}

	return chromosome
}

// Reference returns a sanitized/standardized reference assembly name.
func Reference(reference string) (types.Reference, error) {
	switch reference {
	case "NCBI36", "hg18":
		return types.ReferenceNCBI36, nil
	case "GRCh37", "hg19":
		return types.ReferenceGRCh37, nil
	case "GRCh38", "hg38":
		return types.ReferenceGRCh38, nil
	case "T2T-CHM13v2.0":
		return types.ReferenceTelomereToTelomereV2, nil
	default:
		return "", fmt.Errorf("invalid reference assembly")
	}
}
