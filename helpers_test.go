package main

import "os"

func pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

// nolint: unused
func newInt(i int) *int {
	return &i
}
