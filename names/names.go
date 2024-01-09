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
