package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
	"testing"
)

func TestNewPublicKeyDistributionApi_sortPublishInfo(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	hashes := []types.Hash{mock.Hash(), mock.Hash(), mock.Hash(), mock.Hash(), mock.Hash()}

	pubs := []*PublishInfoState{
		{
			PublishParam: &PublishParam{
				Hash: hashes[0].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address()},
				VerifiedHeight: 0,
				VerifiedStatus: types.PovPublishStatusInit,
				BonusFee:       nil,
				PublishHeight:  100,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[1].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 20,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[2].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 10,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[3].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{mock.Address(), mock.Address(), mock.Address()},
				VerifiedHeight: 10,
				VerifiedStatus: types.PovPublishStatusVerified,
				BonusFee:       nil,
			},
		},
		{
			PublishParam: &PublishParam{
				Hash: hashes[4].String(),
			},
			State: &types.PovPublishState{
				OracleAccounts: []types.Address{},
				VerifiedHeight: 0,
				VerifiedStatus: types.PovPublishStatusInit,
				BonusFee:       nil,
				PublishHeight:  50,
			},
		},
	}

	p := NewPublicKeyDistributionApi(cfgFile, l)
	p.sortPublishInfo(pubs)

	if pubs[0].Hash != hashes[1].String() || pubs[1].Hash != hashes[2].String() || pubs[2].Hash != hashes[3].String() ||
		pubs[3].Hash != hashes[4].String() || pubs[4].Hash != hashes[0].String() {
		t.Fatal()
	}
}
