package types

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type SmartContractBlock struct {
	Address         Address     `msg:"address,extension" json:"address"`
	InternalAccount Address     `msg:"internalAccount,extension" json:"internalAccount"`
	ExtraAddress    []Address   `msg:"extraAddress" json:"extraAddress,omitempty"`
	Owner           Address     `msg:"owner,extension" json:"owner"`
	Abi             ContractAbi `msg:"contract,extension" json:"contract"`
	AbiSchema       string      `msg:"schema" json:"schema"`
	IsUseStorage    bool        `msg:"isUseStorage" json:"isUseStorage"`
	Work            Work        `msg:"work,extension" json:"work"`
	Signature       Signature   `msg:"signature,extension" json:"signature"`
	Version         uint64      `msg:"version" json:"version"`
}

func (sc *SmartContractBlock) GetHash() Hash {
	var extra []byte
	for _, addr := range sc.ExtraAddress {
		extra = append(extra, addr[:]...)
	}
	hash, _ := HashBytes(sc.Address[:], util.String2Bytes(sc.AbiSchema), sc.InternalAccount[:], extra, sc.Owner[:],
		util.Bool2Bytes(sc.IsUseStorage), sc.Abi.AbiHash[:])
	return hash
}

func (sc *SmartContractBlock) Size() int {
	return sc.Msgsize()
}

func (sc *SmartContractBlock) Serialize() ([]byte, error) {
	return sc.MarshalMsg(nil)
}

func (sc *SmartContractBlock) Deserialize(text []byte) error {
	_, err := sc.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

//TODO: improvement
func (sc *SmartContractBlock) IsValid() bool {
	return sc.Work.IsValid(sc.GetHash())
}

func (sc *SmartContractBlock) String() string {
	bytes, _ := json.Marshal(sc)
	return string(bytes)
}
