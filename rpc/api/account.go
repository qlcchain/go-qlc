package api

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type AccountApi struct {
	logger *zap.SugaredLogger
}

func NewAccountApi() *AccountApi {
	return &AccountApi{logger: log.NewLogger("rpc/account")}
}

func (a *AccountApi) Create(seedStr string, i *uint32) (map[string]string, error) {
	var index uint32
	b, err := hex.DecodeString(seedStr)
	if err != nil {
		return nil, err
	}
	if i == nil {
		index = 0
	} else {
		index = *i
	}
	seed, err := types.BytesToSeed(b)
	if err != nil {
		return nil, err
	}
	acc, err := seed.Account(index)
	if err != nil {
		return nil, err
	}
	r := make(map[string]string)
	r["pubKey"] = hex.EncodeToString(acc.Address().Bytes())
	r["privKey"] = hex.EncodeToString(acc.PrivateKey())
	return r, nil
}

func (a *AccountApi) ForPublicKey(pubStr string) (types.Address, error) {
	pub, err := hex.DecodeString(pubStr)
	if err != nil {
		return types.ZeroAddress, err
	}
	addr := types.PubToAddress(pub)
	a.logger.Debug(addr)
	return addr, nil
}

func (a *AccountApi) NewSeed() (string, error) {
	seed, err := types.NewSeed()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(seed[:]), nil
}

type Accounts struct {
	Seed       string `json:"seed"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

func (a *AccountApi) NewAccounts(count *uint32) ([]*Accounts, error) {
	var count1 uint32
	var acs []*Accounts
	if count == nil {
		count1 = 10
	} else {
		count1 = *count
	}
	for i := 0; uint32(i) < count1; i++ {
		seed, err := types.NewSeed()
		if err == nil {
			if a, err := seed.Account(0); err == nil {
				fmt.Println("Seed:", seed.String())
				fmt.Println("Address:", a.Address())
				fmt.Println("Private:", hex.EncodeToString(a.PrivateKey()))
				s := &Accounts{
					Seed:       seed.String(),
					PrivateKey: hex.EncodeToString(a.PrivateKey()),
					PublicKey:  hex.EncodeToString(a.Address().Bytes()),
					Address:    a.Address().String(),
				}
				acs = append(acs, s)
			} else {
				return nil, errors.New("new account error")
			}
		} else {
			return nil, errors.New("new seed error")
		}
	}
	return acs, nil
}

func (a *AccountApi) PublicKey(addr types.Address) string {
	pub := hex.EncodeToString(addr.Bytes())
	return pub
}

func (a *AccountApi) Validate(addr string) bool {
	return types.IsValidHexAddress(addr)
}
