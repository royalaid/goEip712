package eip712

import (
	"fmt"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// type aliases to avoid importing "core" everywhere
type TypedData = apitypes.TypedData
type TypedDataDomain = apitypes.TypedDataDomain
type Types = apitypes.Types
type Type = apitypes.Type
type TypedDataMessage = apitypes.TypedDataMessage

// EncodeForSigning encodes the hash that will be signed for the given EIP712 data
func EncodeForSigning(typedData *TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	return rawData, nil
}
