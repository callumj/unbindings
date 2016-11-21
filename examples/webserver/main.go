package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/callumj/unbindings/core"
)

func handleIt(w http.ResponseWriter, r *http.Request) {
	pth := r.URL.Path
	i, err := core.NewInvocation("./"+strings.TrimPrefix(pth, "/"), false)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}
	var b bytes.Buffer
	go func() {
		scanner := bufio.NewScanner(i.StdOut)
		for scanner.Scan() {
			b.Write(scanner.Bytes())
		}
	}()

	for k, v := range r.Form {
		i.SetOption(k, v[0])
	}

	if err := i.Start(); err != nil {
		log.Panic(err)
	}

	if err := i.Wait(); err != nil {
		log.Panic(err)
	}
	w.Write(b.Bytes())
}

func main() {
	http.HandleFunc("/", handleIt)           // set router
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
