package domain

import (
	"time"
)

type Messages struct {
	From    string
	Message string
	Time    time.Time
	To      string
}

type MessageStorage struct {
	UserNamesAndPass map[string]string
	ToAllMessages    []Messages
	PersonalMessages map[string][]Messages
}
