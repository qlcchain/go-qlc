/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package settlement

import (
	"sort"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func Test_sortInvoiceFun(t *testing.T) {
	var invoices []*InvoiceRecord
	for i := 0; i < 4; i++ {
		s, _ := random.Intn(1000)
		e, _ := random.Intn(1000)
		invoice := &InvoiceRecord{
			StartDate: int64(s),
			EndDate:   int64(e),
		}
		invoices = append(invoices, invoice)
	}

	sort.Slice(invoices, func(i, j int) bool {
		return sortInvoiceFun(invoices[i], invoices[j])
	})
}
