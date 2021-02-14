package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

type BlockDriveInterface interface {
	InitiateUserProfile(ctx contractapi.TransactionContextInterface, name string) (*UserProfile, error)
	ReadUserProfile(ctx contractapi.TransactionContextInterface) (*UserProfile, error)

	CreateDirectory(ctx contractapi.TransactionContextInterface, name string, visibility string) (string, error)
	ReadDirectory(ctx contractapi.TransactionContextInterface, keys string) (*Directory, error)
	ReadDirectories(ctx contractapi.TransactionContextInterface, keys []string) (map[string]*Directory, error)
	AddDirectories(ctx contractapi.TransactionContextInterface, parentKey string, childrenKeys []string) (*Directory, error)
	RemoveDirectories(ctx contractapi.TransactionContextInterface, parentKey string, childrenKeys []string) (*Directory, error)
	RenameDirectory(ctx contractapi.TransactionContextInterface, keys string, name string) (*Directory, error)
	AddFile(ctx contractapi.TransactionContextInterface, key string, files []*FileMeta) (*Directory, error)
	RemoveFile(ctx contractapi.TransactionContextInterface, key string, file []string) (*Directory, error)
	SetDirectoryVisibility(ctx contractapi.TransactionContextInterface, key string, visibility string) (*Directory, error)
	ReadDirectoryHistory(ctx contractapi.TransactionContextInterface, key string) ([]*Directory, error)
	CopyDirectory(ctx contractapi.TransactionContextInterface, source, destination string) error

	AddSubscribers(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error
	AddCooperators(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error
	RemoveSubscribers(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error
	RemoveCooperators(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error

	Subscribe(ctx contractapi.TransactionContextInterface, key string) (*Directory, error)
}
