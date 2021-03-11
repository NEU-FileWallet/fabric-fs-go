package main

import "testing"

var dir = NewDirectory("test", "123", "nmsl", Public, 1231231)

//var dir2 = NewDirectory("test2", "233", "awsl", Public, 123123123)

func TestDirectory_AddCooperators(t *testing.T) {
	dir.AddCooperators([]string{"1"}, []string{"1"})
	for _, cooperator := range dir.Cooperators {
		if cooperator == "1" {
			return
		}
	}
	t.Errorf("fail to add Cooperator")
}

func TestDirectory_AddDirectories(t *testing.T) {
	dirKey := "123123"
	dir.AddDirectories([]string{dirKey})
	for _, directory := range dir.Directories {
		if directory == dirKey {
			return
		}
	}
	t.Errorf("fail to add directories")
}

func TestDirectory_AddSubscribers(t *testing.T) {
	dir.AddSubscribers([]string{"2"}, []string{"2"}, 123)
	for _, subscriber := range dir.Subscribers {
		if subscriber.Id == "2" {
			return
		}
	}
	t.Errorf("fail to add subscriber")
}

func TestDirectory_RemoveCooperators(t *testing.T) {
	dir.RemoveCooperators([]string{"1"})
	for _, cooperator := range dir.Cooperators {
		if cooperator == "1" {
			t.Errorf("fail to remove cooperator")
		}
	}
}

func TestDirectory_RemoveSubscribers(t *testing.T) {
	dir.RemoveSubscribers([]string{"2"})
	for _, subscriber := range dir.Subscribers {
		if subscriber.Id == "2" {
			t.Errorf("fail to remove subscriber")
		}
	}
}

func TestDirectory_IsSubscribers(t *testing.T) {
	dir.AddSubscribers([]string{"1"}, []string{"1"}, 123)
	isSubscriber := dir.IsSubscribers("1", 1)
	if !isSubscriber {
		t.Errorf("should be subscriber but not")
	}
	isSubscriber = dir.IsSubscribers("2", 1)
	if isSubscriber {
		t.Errorf("should not be subscriber")
	}
	isSubscriber = dir.IsSubscribers("1", 124)
	if isSubscriber {
		t.Errorf("subscription expired")
	}
}

func TestDirectory_IsCooperator(t *testing.T) {
	dir.AddCooperators([]string{"2"}, []string{"2"})
	isCooperator := dir.IsCooperator("2")
	if !isCooperator {
		t.Errorf("should be cooperator")
	}
	isCooperator = dir.IsCooperator("3")
	if isCooperator {
		t.Errorf("should not be cooperator")
	}
}

func TestDirectory_IsCreator(t *testing.T) {
	isCreator := dir.IsCreator("123")
	if !isCreator {
		t.Errorf("should be creator")
	}
	isCreator = dir.IsCreator("124")
	if isCreator {
		t.Errorf("should not be creator")
	}
}
