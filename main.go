package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ComputedEvent struct {
	ID       big.Int
	JsonData string
}

const contractABI = `[ABI of OracleContract]`

func main() {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID")
	if err != nil {
		log.Fatal(err)
	}

	// Load the private key
	key, err := keystore.DecryptKey([]byte("YOUR_KEYSTORE_JSON"), "YOUR_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}
	privateKey := key.PrivateKey

	contractAddress := common.HexToAddress("YOUR_CONTRACT_ADDRESS")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			var event ComputedEvent
			err := parsedABI.UnpackIntoInterface(&event, "Computed", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}

			// verify the data
			var astReqDto ComputeRequestDto
			err = json.Unmarshal([]byte(event.JsonData), &astReqDto)
			if err != nil {
				log.Fatal(err)
			}

			result, err := computeRequest(event.JsonData)
			if err != nil {
				log.Fatal(err)
			}

			callContractWriteResult(client, privateKey, contractAddress, event.ID, result)
		}
	}
}

func computeRequest(reqData string) (int, error) {
	url := "http://localhost:3000/compute"

	request, err := http.NewRequest("POST", url, bytes.NewBufferString(reqData))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	var cr ComputeResponse
	err = json.Unmarshal(body, &cr)
	if err != nil {
		var cre ComputeResponseError
		err = json.Unmarshal(body, &cre)
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("compute service return error: %w", cre.Error)
	}
	return cr.Result, nil
}

func callContractWriteResult(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contractAddress common.Address, id big.Int, result int) {
	chainID := big.NewInt(1) // Mainnet chain ID
	nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	contractABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatal(err)
	}

	data, err := contractABI.Pack("writeResult", id, big.NewInt(int64(result)))
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &contractAddress,
		Value:    big.NewInt(0),
		Gas:      uint64(300000),
		GasPrice: gasPrice,
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Result written: %s", signedTx.Hash().Hex())
}
