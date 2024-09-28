package main

import "os"

type wrappedReader func(p []byte) (int, error)

func (actual wrappedReader) Read(p []byte) (int, error) {
	return actual(p)
}

func testHomelabBaseDir() string {
	return "testdata/dummy-base-dir"
}

func pwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func newInt(i int) *int {
	return &i
}
