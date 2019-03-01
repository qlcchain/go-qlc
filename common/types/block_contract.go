package types

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type SmartContractBlock struct {
	Type            BlockType   `msg:"type" json:"type"`
	Address         Address     `msg:"address,extension" json:"address"`
	Previous        Hash        `msg:"previous,extension" json:"previous"`
	InternalAccount Address     `msg:"internalAccount,extension" json:"internalAccount"`
	ExtraAddress    []Address   `msg:"extraAddress" json:"extraAddress,omitempty"`
	Owner           Address     `msg:"owner,extension" json:"owner"`
	Abi             ContractAbi `msg:"contract,extension" json:"contract"`
	IsUseStorage    bool        `msg:"isUseStorage" json:"isUseStorage"`
	Extra           Hash        `msg:"extra,extension" json:"extra"`
	Work            Work        `msg:"work,extension" json:"work"`
	Signature       Signature   `msg:"signature,extension" json:"signature"`
}

func (sc *SmartContractBlock) GetHash() Hash {
	var extra []byte
	for _, addr := range sc.ExtraAddress {
		extra = append(extra, addr[:]...)
	}
	t := []byte{byte(sc.Type)}
	hash, _ := HashBytes(t, sc.Address[:], sc.Previous[:], sc.InternalAccount[:], extra, sc.Owner[:],
		util.Bool2Bytes(sc.IsUseStorage), sc.Abi.AbiHash[:], sc.Extra[:])
	return hash
}

func (sc *SmartContractBlock) Root() Hash {
	return sc.Address.ToHash()
}

func (sc *SmartContractBlock) Size() int {
	return sc.Msgsize()
}

func (sc *SmartContractBlock) GetExtraAddress() []Address {
	return sc.ExtraAddress
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

func (sc *SmartContractBlock) GetType() BlockType {
	return sc.Type
}

//TODO: improvement
func (sc *SmartContractBlock) IsValid() bool {
	return sc.Work.IsValid(sc.Address.ToHash())
}

func (sc *SmartContractBlock) String() string {
	bytes, _ := json.Marshal(sc)
	return string(bytes)
}
