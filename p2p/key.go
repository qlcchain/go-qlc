package p2p

import (
	"encoding/base64"
	"io/ioutil"
	mrand "math/rand"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
)

// LoadNetworkKeyFromFile load network priv key from file.
func loadNetworkKeyFromFile(path string) (crypto.PrivKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		key, err := generateEd25519Key()
		if err != nil {
			return nil, err
		}

		str, _ := marshalNetworkKey(key)

		ioutil.WriteFile(path, []byte(str), 0600)

		return key, err
	}
	return unmarshalNetworkKey(string(data))
}

// LoadNetworkKeyFromFileOrCreateNew load network priv key from file or create new one.
func LoadNetworkKeyFromFileOrCreateNew(path string) (crypto.PrivKey, error) {
	if path == "" {
		return generateEd25519Key()
	}
	return loadNetworkKeyFromFile(path)
}

// UnmarshalNetworkKey unmarshal network key.
func unmarshalNetworkKey(data string) (crypto.PrivKey, error) {
	binaryData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(binaryData)
}

// MarshalNetworkKey marshal network key.
func marshalNetworkKey(key crypto.PrivKey) (string, error) {
	binaryData, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(binaryData), nil
}

// GenerateEd25519Key return a new generated Ed22519 Private key.
func generateEd25519Key() (crypto.PrivKey, error) {
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	key, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	return key, err
}
