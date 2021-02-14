package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

const validity = 315532800

//SmartContract The smart contract implementation.
type SmartContract struct {
	contractapi.Contract
}

func iteration(ctx contractapi.TransactionContextInterface, sourceDirKey, destinationDirKey string, creatorID, creatorName string, timestamp int64) error {

	sourceDir, err := getDirectory(ctx, sourceDirKey, Subscriber)
	if err != nil {
		return err
	}

	destinationDir, err := getDirectory(ctx, destinationDirKey, Subscriber)
	if err != nil {
		return err
	}

	cloneDir := NewDirectory(sourceDir.Name, creatorID, creatorName, sourceDir.Visibility, timestamp)
	cloneDirKey := CalculateDirectoryKey(timestamp, creatorID, sourceDir.Name)
	log.Println("cloneDirKey")
	log.Println(cloneDirKey)
	cloneDir.Files = sourceDir.Files

	for _, dirKey := range sourceDir.Directories {
		if err := iteration(ctx, dirKey, cloneDirKey, creatorID, creatorName, timestamp); err != nil {
			return err
		}
	}

	destinationDir.AddDirectories([]string{cloneDirKey})
	cloneDir.Cooperators = destinationDir.Cooperators
	cloneDir.Subscribers = destinationDir.Subscribers
	cloneDir.IDNameMap = destinationDir.IDNameMap
	err = PutJsonState(ctx, cloneDirKey, cloneDir)
	if err != nil {
		return err
	}
	return PutJsonState(ctx, destinationDirKey, destinationDir)
}

func (s *SmartContract) CopyDirectory(ctx contractapi.TransactionContextInterface, source, destination string) error {
	id, err := getUserID(ctx)
	if err != nil {
		return err
	}
	userProfile, err := getUserProfile(ctx, id)
	if err != nil {
		return err
	}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}

	err = iteration(ctx, source, destination, id, userProfile.Name, timestamp.Seconds)
	return err
}

//RemoveFile Remove file from directory. It will return an updated directory or an error.
func (s *SmartContract) RemoveFile(ctx contractapi.TransactionContextInterface, key string, file []string) (*Directory, error) {
	directory, err := getDirectory(ctx, key, Cooperator)
	if err != nil {
		return nil, err
	}

	directory.RemoveFiles(file)

	if err = directory.Save(ctx, key); err != nil {
		return nil, err
	}

	return directory, nil
}

//InitiateUserProfile initiate the profile of specific user. If the user profile existed, it will return this user's profile.
func (s *SmartContract) InitiateUserProfile(ctx contractapi.TransactionContextInterface, name string) (*UserProfile, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	timestamp, _ := ctx.GetStub().GetTxTimestamp()

	bytes, _ := ctx.GetStub().GetState(id)
	if len(bytes) == 0 {

		privateFolder := NewDirectory("All Files", id, name, "Private", timestamp.Seconds)
		privateFolderKey := CalculateDirectoryKey(timestamp.Seconds, id, "All Files")

		shareFolder := NewDirectory("Share", id, name, "Private", timestamp.Seconds)
		shareFolderKey := CalculateDirectoryKey(timestamp.Seconds, id, "Share")

		subscriptionFolder := NewDirectory("Subscription", id, name, "Private", timestamp.Seconds)
		subscriptionFolderKey := CalculateDirectoryKey(timestamp.Seconds, id, "Subscription")

		privateFolder.Directories = []string{shareFolderKey, subscriptionFolderKey}

		if err = shareFolder.Save(ctx, shareFolderKey); err != nil {
			return nil, err
		}
		if err = subscriptionFolder.Save(ctx, subscriptionFolderKey); err != nil {
			return nil, err
		}
		if err = privateFolder.Save(ctx, privateFolderKey); err != nil {
			return nil, err
		}

		profile := &UserProfile{
			Id:      id,
			Name:    name,
			Private: privateFolderKey,
			//Share:         shareFolderKey,
			//Subscriptions: subscriptionFolderKey,
		}

		profileValue, _ := json.Marshal(profile)
		if err = ctx.GetStub().PutState(id, profileValue); err != nil {
			return nil, err
		}

		return profile, nil
	}

	profile := new(UserProfile)
	if err = json.Unmarshal(bytes, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *SmartContract) ReadUserProfile(ctx contractapi.TransactionContextInterface) (*UserProfile, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	bytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	userProle := new(UserProfile)
	if err = json.Unmarshal(bytes, userProle); err != nil {
		return nil, err
	}

	return userProle, nil
}

func (s *SmartContract) ReadUserName(ctx contractapi.TransactionContextInterface, userID string) (string, error) {
	userProfile := new(UserProfile)
	if err := GetJsonState(ctx, userID, userProfile); err != nil {
		return "", err
	}
	return userProfile.Name, nil
}

func (s *SmartContract) ReadDirectories(ctx contractapi.TransactionContextInterface, keys []string) (map[string]*Directory, error) {
	resultMap := make(map[string]*Directory)
	for _, key := range keys {
		directory, err := getDirectory(ctx, key, Subscriber)
		if err != nil {
			continue
		}
		resultMap[key] = directory
	}
	return resultMap, nil
}

func (s *SmartContract) ReadDirectory(ctx contractapi.TransactionContextInterface, key string) (*Directory, error) {
	return getDirectory(ctx, key, Subscriber)
}

func (s *SmartContract) AddDirectories(ctx contractapi.TransactionContextInterface, parentKey string, newDireKeys []string) (*Directory, error) {
	directory, err := getDirectory(ctx, parentKey, Cooperator)
	if err != nil {
		return nil, err
	}

	newDirs, err := s.ReadDirectories(ctx, newDireKeys)
	children, err := s.ReadDirectories(ctx, directory.Directories)
	newDirsNames := make([]string, 0)
	childrenNames := make([]string, 0)
	for _, newDir := range newDirs {
		newDirsNames = append(newDirsNames, newDir.Name)
	}
	for _, childrenDir := range children {
		childrenNames = append(childrenNames, childrenDir.Name)
	}
	intersection := getIntersection(newDirsNames, childrenNames)
	if len(intersection) > 0 {
		return nil, fmt.Errorf("directory name conflict")
	}

	directory.AddDirectories(newDireKeys)
	if err = directory.Save(ctx, parentKey); err != nil {
		return nil, err
	}

	return directory, nil
}

func (s *SmartContract) RemoveDirectories(ctx contractapi.TransactionContextInterface, parentKey string, childrenKeys []string) (*Directory, error) {
	directory, err := getDirectory(ctx, parentKey, Cooperator)
	if err != nil {
		return nil, err
	}

	directory.RemoveDirectories(childrenKeys)
	if err = directory.Save(ctx, parentKey); err != nil {
		return nil, err
	}

	return directory, nil
}

func (s *SmartContract) RenameDirectory(ctx contractapi.TransactionContextInterface, key string, name string) (*Directory, error) {
	directory, err := getDirectory(ctx, key, Cooperator)
	if err != nil {
		return nil, err
	}

	directory.Name = name
	if err = directory.Save(ctx, key); err != nil {
		return nil, err
	}

	return directory, nil
}

func (s *SmartContract) AddFile(ctx contractapi.TransactionContextInterface, key string, files []*FileMeta) (*Directory, error) {
	directory, err := getDirectory(ctx, key, Cooperator)
	if err != nil {
		return nil, err
	}

	directory.AddFiles(files)
	if err = directory.Save(ctx, key); err != nil {
		return nil, err
	}

	return directory, nil
}

func (s *SmartContract) CreateDirectory(ctx contractapi.TransactionContextInterface, name string, visibility string) (string, error) {
	creatorID, err := getUserID(ctx)
	if err != nil {
		return "", err
	}
	creatorName, err := s.ReadUserName(ctx, creatorID)
	if err != nil {
		return "", err
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", err
	}

	directory := NewDirectory(name, creatorID, creatorName, visibility, timestamp.Seconds)
	key := CalculateDirectoryKey(timestamp.Seconds, creatorID, name)

	if err = directory.Save(ctx, key); err != nil {
		return "", err
	}

	return key, nil
}

func (s *SmartContract) SetDirectoryVisibility(ctx contractapi.TransactionContextInterface, key string, visibility string) (*Directory, error) {
	directory, err := getDirectory(ctx, key, Cooperator)
	if err != nil {
		return nil, err
	}
	directory.Visibility = visibility
	if err = directory.Save(ctx, key); err != nil {
		return nil, err
	}
	return directory, nil
}

func (s *SmartContract) ReadDirectoryHistory(ctx contractapi.TransactionContextInterface, key string) ([]*Directory, error) {
	_, err := getDirectory(ctx, key, Subscriber)
	if err != nil {
		return nil, err
	}

	iterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}

	dirs := make([]*Directory, 0)

	for iterator.HasNext() {
		mod, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		dir := new(Directory)
		if err = json.Unmarshal(mod.GetValue(), dir); err != nil {
			return nil, err
		}
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

func getNameByID(ctx contractapi.TransactionContextInterface, ids []string) ([]string, error) {
	names := make([]string, len(ids))

	for index, id := range ids {
		bytes, err := ctx.GetStub().GetState(id)
		if err != nil {
			return nil, err
		}
		userProle := new(UserProfile)
		if err = json.Unmarshal(bytes, userProle); err != nil {
			return nil, err
		}
		names[index] = userProle.Name
	}

	return names, nil
}

type Action = func(directory *Directory, ids []string, names []string, timestamp int64)

func updateDirectoryAccess(
	ctx contractapi.TransactionContextInterface,
	key string,
	ids []string,
	recursive bool,
	action Action,
) error {
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	names, err := getNameByID(ctx, ids)
	if err != nil {
		return err
	}

	return updateIteration(ctx, key, ids, names, timestamp.Seconds, recursive, action)
}

func updateIteration(ctx contractapi.TransactionContextInterface, dirKey string, ids, names []string, timestamp int64, recursive bool, action Action) error {
	dir, err := getDirectory(ctx, dirKey, All)
	if err != nil {
		return err
	}
	action(dir, ids, names, timestamp)
	if err = dir.Save(ctx, dirKey); err != nil {
		return err
	}
	if recursive {
		for _, dirKey := range dir.Directories {
			err = updateIteration(ctx, dirKey, ids, names, timestamp, recursive, action)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SmartContract) AddSubscribers(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error {
	return updateDirectoryAccess(ctx, key, ids, recursive, func(directory *Directory, ids []string, names []string, timestamp int64) {
		directory.AddSubscribers(ids, names, timestamp+validity)
	})
}

func (s *SmartContract) AddCooperators(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error {
	return updateDirectoryAccess(ctx, key, ids, recursive, func(directory *Directory, ids []string, names []string, timestamp int64) {
		directory.AddCooperators(ids, names)
	})
}

func (s *SmartContract) RemoveSubscribers(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error {
	return updateDirectoryAccess(ctx, key, ids, recursive, func(directory *Directory, ids []string, names []string, timestamp int64) {
		directory.RemoveSubscribers(ids)
	})
}

func (s *SmartContract) RemoveCooperators(ctx contractapi.TransactionContextInterface, key string, ids []string, recursive bool) error {
	return updateDirectoryAccess(ctx, key, ids, recursive, func(directory *Directory, ids []string, names []string, timestamp int64) {
		directory.RemoveCooperators(ids)
	})
}

func (s *SmartContract) Subscribe(ctx contractapi.TransactionContextInterface, key string) (*Directory, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	name, err := s.ReadUserName(ctx, id)
	if err != nil {
		return nil, err
	}

	ids := []string{id}
	names := []string{name}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return nil, err
	}

	directory, err := getDirectory(ctx, key, All)
	if err != nil {
		return nil, err
	}

	if directory.IsSubscribers(id, timestamp.Seconds) {
		return directory, nil
	}

	if directory.IsCreator(id) || directory.IsCooperator(id) || directory.Visibility == Public {
		directory.AddSubscribers(ids, names, timestamp.Seconds+validity)
	} else {
		return nil, fmt.Errorf("can't access private directory")
	}

	err = directory.Save(ctx, key)
	if err != nil {
		return nil, err
	}
	return directory, nil
}
