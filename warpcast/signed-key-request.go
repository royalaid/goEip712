package warpcast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type WarpcastBody struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	RequestFid int64  `json:"requestFid"`
	Deadline   int64  `json:"deadline"`
	Signature  string `json:"signature"`
}

type SignedKeyRequest struct {
	Token       string `json:"token"`
	DeeplinkUrl string `json:"deeplinkUrl"`
	Key         string `json:"key"`
	RequestFid  int    `json:"requestFid"`
	State       string `json:"state"`
}

type SignedKeyRequestResponse struct {
	Result struct {
		SignedKeyRequest SignedKeyRequest `json:"signedKeyRequest"`
	}
}

func RequestToken(payload WarpcastBody) (*SignedKeyRequestResponse, error) {
	jsonBody, err := json.MarshalIndent(payload, "", "  ")
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.warpcast.com/v2/signed-key-requests", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	var signedKeyRequest SignedKeyRequestResponse
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("%s", jsonBody)
	// Print the HTTP status code
	log.Println("Response status:", resp.Status)
	err = json.NewDecoder(resp.Body).Decode(&signedKeyRequest)
	if err != nil {
		return nil, err
	}
	return &signedKeyRequest, nil
}

func CheckTokenStatus(signedKeyRequest SignedKeyRequestResponse) ([]byte, error) {
	client := &http.Client{}
	statusReq, err := http.NewRequest("GET", fmt.Sprintf("https://api.warpcast.com/v2/signed-key-request?token=%s", signedKeyRequest.Result.SignedKeyRequest.Token), nil)
	if err != nil {
		return nil, err
	}

	statusReq.Header.Set("Content-Type", "application/json")
	statusResp, err := client.Do(statusReq)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(statusResp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, err

}
