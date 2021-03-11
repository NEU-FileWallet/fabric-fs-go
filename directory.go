package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

type Directory struct {
	Name        string            `json:"name"`
	Directories []string          `json:"directories"`
	Files       []*FileMeta       `json:"files"`
	Creator     string            `json:"creator"`
	Editor      string            `json:"editor"`
	Date        int64             `json:"date"`
	Cooperators []string          `json:"cooperators"`
	Subscribers []*SubscriberMeta `json:"subscribers"`
	Deleted     bool              `json:"deleted"`
	IDNameMap   map[string]string `json:"idNameMap"`
	Visibility  string            `json:"visibility"`
}

type Privilege int

const (
	All Privilege = iota
	Subscriber
	Cooperator
)

const (
	Public  = "Public"
	Private = "Private"
)

var privilegeError = fmt.Errorf("illegal access")

func getDirectory(ctx contractapi.TransactionContextInterface, key string) (*Directory, error) {
	directory := new(Directory)
	if err := GetJsonState(ctx, key, directory); err != nil {
		return nil, fmt.Errorf("directory doesn't exist")
	}
	return directory, nil
}

//func getDirectoryWithPrivilege(ctx contractapi.TransactionContextInterface, key string, privilege Privilege) (*Directory, error) {
//	id, err := getUserID(ctx)
//	if err != nil {
//		return nil, err
//	}
//	timestamp, err := ctx.GetStub().GetTxTimestamp()
//	if err != nil {
//		return nil, err
//	}
//
//    directory, err := getDirectory(ctx, key)
//	if err != nil {
//        return nil, err
//    }
//
//	privilegeError := fmt.Errorf("illegal access")
//	switch privilege {
//	case Subscriber:
//		if !directory.IsCooperator(id) && !directory.IsSubscribers(id, timestamp.Seconds) {
//			return nil, privilegeError
//		}
//	case Cooperator:
//		if !directory.IsCooperator(id) {
//			return nil, privilegeError
//		}
//	}
//
//	return directory, nil
//}

func CalculateDirectoryKey(timestamp int64, id, name string) string {
	return SHA256(fmt.Sprintf("%s%d%s", id, timestamp, name))
}

func NewDirectory(name, creatorID, creatorName string, visibility string, date int64) *Directory {
	return &Directory{
		Name:        name,
		Directories: make([]string, 0),
		Files:       make([]*FileMeta, 0),
		Creator:     creatorID,
		Date:        date,
		Editor:      creatorID,
		Cooperators: []string{creatorID},
		Subscribers: make([]*SubscriberMeta, 0),
		Deleted:     false,
		IDNameMap:   map[string]string{creatorID: creatorName},
		Visibility:  visibility,
	}
}

func (d *Directory) CheckPrivilege(ctx contractapi.TransactionContextInterface, privilege Privilege) (bool, error) {
	id, err := getUserID(ctx)
	if err != nil {
		return false, err
	}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return false, err
	}

	switch privilege {
	case Subscriber:
		if !d.IsCooperator(id) && !d.IsSubscribers(id, timestamp.Seconds) {
			return false, nil
		}
	case Cooperator:
		if !d.IsCooperator(id) {
			return false, nil
		}
	}
	return true, nil
}

func (d *Directory) ToString() string {
	bytes, _ := json.Marshal(d)
	return string(bytes)
}

func (d *Directory) IsCreator(id string) bool {
	return id == d.Creator
}

func (d *Directory) IsCooperator(id string) bool {
	for _, cooperator := range d.Cooperators {
		if cooperator == id {
			return true
		}
	}
	return false
}

func (d *Directory) IsSubscribers(id string, timestamp int64) bool {
	for _, subscriber := range d.Subscribers {
		if subscriber.Id == id && subscriber.DueDate > timestamp {
			return true
		}
	}
	return false
}

func (d *Directory) AddIDNameMap(id []string, names []string) {
	for index, item := range id {
		d.IDNameMap[item] = names[index]
	}
}

func (d *Directory) RemoveIDNameMap(id []string) {
	for _, i := range id {
		delete(d.IDNameMap, i)
	}
}

func (d *Directory) AddCooperators(ids []string, names []string) {
	log.Println("AddCooperators:")
	log.Println(names)
	for _, id := range ids {
		if !d.IsCooperator(id) {
			d.Cooperators = append(d.Cooperators, id)
		}
	}

	d.AddIDNameMap(ids, names)
}

func (d *Directory) RemoveCooperators(id []string) {
	record := make(map[string]bool)
	remains := make([]string, 0)
	for _, i := range id {
		record[i] = true
	}
	for _, i := range d.Cooperators {
		if record[i] {
			continue
		}
		remains = append(remains, i)
	}
	d.Cooperators = remains
}

func (d *Directory) AddSubscribers(ids []string, names []string, date int64) {
	newSubscriberArray := make([]*SubscriberMeta, 0)
	existedSubscriberMap := make(map[string]bool)

	for _, subscriber := range d.Subscribers {
		if subscriber.DueDate > date {
			existedSubscriberMap[subscriber.Id] = true
		}
	}

	for _, id := range ids {
		if existedSubscriberMap[id] {
			continue
		}

		newSubscriberArray = append(newSubscriberArray, &SubscriberMeta{
			Id:      id,
			DueDate: date,
		})
	}
	d.Subscribers = append(d.Subscribers, newSubscriberArray...)

	for index, item := range ids {
		d.IDNameMap[item] = names[index]
	}
}

func (d *Directory) RemoveSubscribers(id []string) {
	record := make(map[string]bool)
	remains := make([]*SubscriberMeta, 0)
	for _, i := range id {
		record[i] = true
	}

	for _, i := range d.Subscribers {
		if record[i.Id] {
			continue
		}
		remains = append(remains, i)
	}

	d.Subscribers = remains
}

func (d *Directory) AddDirectories(keys []string) {
	d.Directories = append(d.Directories, keys...)
}

func (d *Directory) RemoveDirectories(keys []string) {
	record := make(map[string]bool)
	remains := make([]string, 0)
	for _, i := range keys {
		record[i] = true
	}

	for _, i := range d.Directories {
		if record[i] {
			continue
		}
		remains = append(remains, i)
	}

	d.Directories = remains
}

func (d *Directory) AddFiles(fileMetas []*FileMeta) {
	temp := make(map[string]bool)
	for _, file := range d.Files {
		temp[file.Name] = true
	}

	for _, meta := range fileMetas {
		if temp[meta.Name] {
			meta.Name = meta.Name + uuid.New().String()[0:4]
		}
	}

	d.Files = append(d.Files, fileMetas...)
}

func (d *Directory) RemoveFiles(names []string) {
	record := make(map[string]bool)
	remains := make([]*FileMeta, 0)
	for _, i := range names {
		record[i] = true
	}

	for _, i := range d.Files {
		if record[i.Name] {
			continue
		}

		remains = append(remains, i)
	}

	d.Files = remains
}

func (d *Directory) Save(ctx contractapi.TransactionContextInterface, key string) error {
	var err error
	d.Editor, err = getUserID(ctx)
	if err != nil {
		return err
	}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	d.Date = timestamp.Seconds
	if err := PutJsonState(ctx, key, d); err != nil {
		return err
	}
	return nil
}
