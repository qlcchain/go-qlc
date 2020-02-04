package dpki

import (
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/types"
)

const (
	// following prefix should be unique in DPKI contract, not global
	PovContractStatePrefixPKDPS = byte(1) // publish state
	PovContractStatePrefixPKDVS = byte(2) // verifier state
)

func PovSetPublishState(csdb *statedb.PovContractStateDB, rawKey []byte, ps *types.PovPublishState) error {
	trieKey := types.PovCreateContractLocalStateKey(PovContractStatePrefixPKDPS, rawKey)

	val, err := ps.Serialize()
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, val)
}

func PovGetPublishState(csdb *statedb.PovContractStateDB, rawKey []byte) (*types.PovPublishState, error) {
	trieKey := types.PovCreateContractLocalStateKey(PovContractStatePrefixPKDPS, rawKey)

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}
	if len(valBytes) == 0 {
		return nil, errors.New("key not exist in ")
	}

	ps := types.NewPovPublishState()
	err = ps.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize publish state err %s", err)
	}

	return ps, nil
}

func PovSetVerifierState(csdb *statedb.PovContractStateDB, rawKey []byte, ps *types.PovVerifierState) error {
	trieKey := types.PovCreateContractLocalStateKey(PovContractStatePrefixPKDVS, rawKey)

	val, err := ps.Serialize()
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, val)
}

func PovGetVerifierState(csdb *statedb.PovContractStateDB, rawKey []byte) (*types.PovVerifierState, error) {
	trieKey := types.PovCreateContractLocalStateKey(PovContractStatePrefixPKDVS, rawKey)

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil {
		return nil, err
	}
	if len(valBytes) == 0 {
		return nil, errors.New("get empty value")
	}

	vs := types.NewPovVerifierState()
	err = vs.Deserialize(valBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize verifier state err %s", err)
	}

	return vs, nil
}
