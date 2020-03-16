/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/hex"

	"github.com/qlcchain/go-qlc/common/types"
)

var (
	// qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq
	priv1, _ = hex.DecodeString("7098c089e66bd66476e3b88df8699bcd4dacdd5e1e5b41b3c598a8a36d851184d992a03b7326b7041f689ae727292d761b329a960f3e4335e0a7dcf2c43c4bcf")
	// qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te
	priv2, _ = hex.DecodeString("31ee4e16826569dc631b969e71bd4c46d5c0df0daeca6933f46586f36f49537cd929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393")
	account1 = types.NewAccount(priv1)
	account2 = types.NewAccount(priv2)
)
