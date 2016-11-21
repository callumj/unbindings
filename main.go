package main

import (
	"log"
	"strconv"
	"time"

	"github.com/callumj/unbindings/core"
)

func main() {
	i, err := core.NewInvocation("ruby examples/resident.rb", true)
	if err != nil {
		log.Panic(err)
	}
	opts := map[string]string{
		"Foo":   "Bar does this work for you?",
		"Bar":   "[1,2,3,4]",
		"Thing": "ABC=true",
	}

	for k, v := range opts {
		i.SetOption(k, v)
	}

	if err := i.Start(); err != nil {
		log.Panic(err)
	}

	go func() {
		c := time.Tick(2 * time.Second)
		for now := range c {
			i.SetOption("time", strconv.Itoa(now.Second()))
		}
	}()

	if err := i.Wait(); err != nil {
		log.Panic(err)
	}
}
