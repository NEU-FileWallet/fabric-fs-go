package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

func main() {
	code, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := code.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
