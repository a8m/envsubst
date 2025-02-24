package main

import (
	"fmt"
	"log"

	"github.com/a8m/envsubst"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Env    string
	Host   string
	Region string
}

func main() {
	data, err := envsubst.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("envsubst error: %v", err)
	}
	c := new(Config)
	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Fatalf("yaml error: %v", err)
	}
	fmt.Println(c.Env, c.Host, c.Region) // dev, localhost, us-east-1
}
