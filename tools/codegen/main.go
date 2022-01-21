package main

import (
	"io/ioutil"
	"path/filepath"
)

//go:generate go run .

func main() {
	codegen := codeGen{}
	v1Config := codegen.createV1Config()
	err := ioutil.WriteFile(filepath.Join("..", "..", "pkg", "integrations", "v1", "config.go"), []byte(v1Config), 0664)
	if err != nil {
		panic(err)
	}

	v2Config := codegen.createV2Config()
	err = ioutil.WriteFile(filepath.Join("..", "..", "pkg", "integrations", "v2", "config.go"), []byte(v2Config), 0664)
	if err != nil {
		panic(err)
	}
}
