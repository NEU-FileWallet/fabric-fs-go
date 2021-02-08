package main

type FileMeta struct {
	Cid        string `json:"cid"`
	CreateDate int64  `json:"createDate"`
	Name       string `json:"name"`
	Key        string `json:"key"`
}
