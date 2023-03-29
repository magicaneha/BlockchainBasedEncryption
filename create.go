/*
@Neha Shecter
*/
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	MetaData  string    `json:"metaData"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserContract struct {
	contractapi.Contract
}

func (uc *UserContract) RegisterUser(ctx contractapi.TransactionContextInterface, userID string, name string, email string, metaData string) error {
	user := User{
		ID:        userID,
		Name:      name,
		Email:     email,
		MetaData:  metaData,
		CreatedAt: time.Now(),
	}

	userAsBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	if userAsBytes != nil {
		return fmt.Errorf("user already exists")
	}

	userAsBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err)
	}

	err = ctx.GetStub().PutState(userID, userAsBytes)
	if err != nil {
		return fmt.Errorf("failed to write to world state: %v", err)
	}

	return nil
}

func (uc *UserContract) GetUser(ctx contractapi.TransactionContextInterface, userID string) (*User, error) {
	userAsBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("user does not exist")
	}

	var user User
	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return &user, nil
}

func main() {
	userContract := new(UserContract)

	cc, err := contractapi.NewChaincode(userContract)
	if err != nil {
		fmt.Printf("Error creating user chaincode: %v", err)
		return
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting user chaincode: %v", err)
	}
}
