# httpflow

[![Build Status](https://travis-ci.org/karupanerura/go-httpflow.svg?branch=master)](https://travis-ci.org/karupanerura/go-httpflow)
[![codecov](https://codecov.io/gh/karupanerura/go-httpflow/branch/master/graph/badge.svg)](https://codecov.io/gh/karupanerura/go-httpflow)
[![GoDoc](https://godoc.org/github.com/karupanerura/go-httpflow?status.svg)](http://godoc.org/github.com/karupanerura/go-httpflow)

Simple web api client framework for go programming language.

Example:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/k0kubun/pp"
	httpflow "github.com/karupanerura/go-httpflow"
)

type User struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
}

type UsersGetInstancesSession struct {
	httpflow.NobodyRequestBuilder
	httpflow.JsonResponseHandler
}

func NewUsersGetInstancesSession(id int) *UsersGetInstancesSession {
	netURL, err := url.Parse("https://jsonplaceholder.typicode.com/users/" + strconv.Itoa(id))
	if err != nil {
		panic(err)
	}

	return &UsersGetInstancesSession{
		NobodyRequestBuilder: httpflow.NobodyRequestBuilder{
			RequestMethod: http.MethodGet,
			RequestURL:    netURL,
		},
		JsonResponseHandler: httpflow.JsonResponseHandler{},
	}
}

func (r *UsersGetInstancesSession) ParseBody() (*User, error) {
	var body User
	err := r.DecodeJSON(&body)
	if err != nil {
		return nil, err
	}

	return &body, err
}

func main() {
	agent := httpflow.NewAgent(http.DefaultClient)
	session := NewUsersGetInstancesSession(1)
	err := agent.RunSession(session)
	if err != nil {
		log.Fatal(err)
	}

	res, err := session.ParseBody()
	if err != nil {
		log.Fatal(err)
	}

	pp.Print(res)
}
```
