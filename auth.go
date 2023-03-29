package main

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AuthenticationContract struct {
	contractapi.Contract
}

type Authentication struct {
	UserID      string    `json:"userID"`
	AuthTime    time.Time `json:"authTime"`
}

func (ac *AuthenticationContract) Authenticate(ctx contractapi.TransactionContextInterface, userID string) error {
	auth := Authentication{
		UserID:   userID,
		AuthTime: time.Now(),
	}

	authAsBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	if authAsBytes != nil {
		return fmt.Errorf("user already authenticated")
	}

	authAsBytes, err = json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("failed to marshal authentication data: %v", err)
	}

	err = ctx.GetStub().PutState(userID, authAsBytes)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

func main() {
	authenticationContract := new(AuthenticationContract)

	cc, err := contractapi.NewChaincode(authenticationContract)
	if err != nil {
		fmt.Printf("Error creating authentication chaincode: %v", err)
		return
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting authentication chaincode: %v", err)
	}
}
