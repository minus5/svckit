package example

import "time"

//go:generate go run generate.go

type Person struct {
	Name         string
	Age          int            `json:"a"`
	Children     map[int]*Child `json:"c"`
	NotInAdapter string         `adp:"-"`
}

type Child struct {
	DateOfBirth  time.Time `json:"dob"`
	NotInAdapter string    `adp:"-"`
}
