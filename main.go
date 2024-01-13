package main

import (
	"encoding/json"
	"fmt"
	"goEip712/eip712"
	"goEip712/warpcast"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/joho/godotenv"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func submitTokenPayload() (*warpcast.SignedKeyRequest, error) {

	//Setup Wallet
	mnemonic := os.Getenv("APP_MNEMONIC")
	if mnemonic == "" {
		return nil, fmt.Errorf("No mnemonic provided")
	}

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	account, err := wallet.Derive(accounts.DefaultBaseDerivationPath, true)
	if err != nil {
		return nil, err
	}

	var address = account.Address.Hex()
	key, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}
	signer := NewDefaultSigner(key)

	//Get other needed variables
	appFid, err := strconv.ParseInt(os.Getenv("APP_FID"), 10, 64)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Unix()

	var eipBody = &eip712.TypedData{
		Domain: eip712.TypedDataDomain{
			Name:              "Farcaster SignedKeyRequestValidator",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(10),
			VerifyingContract: "0x00000000fc700472606ed4fa22623acf62c60553",
		},
		Types: eip712.Types{
			"EIP712Domain": {
				{
					Name: "name",
					Type: "string",
				},
				{
					Name: "version",
					Type: "string",
				},
				{
					Name: "chainId",
					Type: "uint256",
				},
				{
					Name: "verifyingContract",
					Type: "address",
				},
			},
			"SignedKeyRequest": {
				{Name: "requestFid", Type: "uint256"},
				{Name: "key", Type: "bytes"},
				{Name: "deadline", Type: "uint256"},
			},
		},
		Message: eip712.TypedDataMessage{
			"requestFid": fmt.Sprintf("%d", appFid),
			"key":        address,
			"deadline":   fmt.Sprintf("%d", deadline),
		},
		PrimaryType: "SignedKeyRequest",
	}

	signedTx, err := signer.SignTypedData(eipBody)

	token, err := warpcast.RequestToken(
		warpcast.WarpcastBody{
			Key:        address,
			Name:       "Test",
			RequestFid: appFid,
			Deadline:   deadline,
			Signature:  fmt.Sprintf("0x%x", signedTx),
		})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &token.Result.SignedKeyRequest, nil

}

func main() {

	err := godotenv.Load()
	if err != nil {
		return
	}

	http.HandleFunc("/signer/", func(w http.ResponseWriter, r *http.Request) {
		signedKeyRequest, err := submitTokenPayload()
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(signedKeyRequest)
		if err != nil {
			log.Printf("An error occured: %s\n", err.Error())
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
