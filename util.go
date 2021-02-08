package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func SHA256(message string) string {
	return SHA256Bytes([]byte(message))
}

func SHA256Bytes(bytes []byte) string {
	hash := sha256.New()
	hash.Write(bytes)
	hashBytes := hash.Sum(nil)
	hashText := hex.EncodeToString(hashBytes)
	return hashText
}

func GetJsonState(ctx contractapi.TransactionContextInterface, key string, variable interface{}) error {
	bytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, variable)
	return err
}

func PutJsonState(ctx contractapi.TransactionContextInterface, key string, variable interface{}) error {
	bytes, err := json.Marshal(variable)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(key, bytes)
}

func getUserID(ctx contractapi.TransactionContextInterface) (string, error) {
	cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return "", err
	}

	return SHA256Bytes(cert.Raw)[0:8], nil
}

func getUserProfile(ctx contractapi.TransactionContextInterface, id string) (*UserProfile, error) {
	userProfile := new(UserProfile)
	err := GetJsonState(ctx, id, userProfile)
	if err != nil {
		return nil, err
	}
	return userProfile, nil
}

func getIntersection(strings1, strings2 []string) []string {
	temp := make(map[string]bool)
	intersection := make([]string, 0)
	for _, str := range strings1 {
		temp[str] = true
	}
	for _, str := range strings2 {
		if temp[str] {
			intersection = append(intersection, str)
		}
	}
	return intersection
}
