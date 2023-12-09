package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
	"goEip712/eip712"
)

type defaultSigner struct {
	key *ecdsa.PrivateKey
}

type Signer interface {
	Sign(data []byte) ([]byte, error)
	SignTypedData(typedData *eip712.TypedData) ([]byte, error)
}

func (d *defaultSigner) sign(sighash []byte, isCompressedKey bool) ([]byte, error) {
	signature, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(d.key), sighash, false)
	if err != nil {
		return nil, err
	}

	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}

// addEthereumPrefix adds the ethereum prefix to the data.
func addEthereumPrefix(data []byte) []byte {
	return []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
}

// hashWithEthereumPrefix returns the hash that should be signed for the given data.
func hashWithEthereumPrefix(data []byte) []byte {
	return crypto.Keccak256(addEthereumPrefix(data))
}

func (d *defaultSigner) Sign(data []byte) (signature []byte, err error) {
	hash := hashWithEthereumPrefix(data)
	if err != nil {
		return nil, err
	}

	return d.sign(hash, true)
}

func (d *defaultSigner) SignTypedData(typedData *eip712.TypedData) ([]byte, error) {
	rawData, err := eip712.EncodeForSigning(typedData)
	if err != nil {
		return nil, err
	}

	sighash := crypto.Keccak256(rawData)

	return d.sign(sighash, false)
}

func NewDefaultSigner(key *ecdsa.PrivateKey) Signer {
	return &defaultSigner{
		key: key,
	}
}
