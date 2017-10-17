# gotcha

Simple web api client builder for go programming language.

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
	"github.com/karupanerura/gotcha"
)

type User struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
}

type UsersGetInstancesSession struct {
	gotcha.NobodyRequestBuilder
	gotcha.JsonResponseHandler
}

func NewUsersGetInstancesSession(id int) *UsersGetInstancesSession {
	netURL, err := url.Parse("https://jsonplaceholder.typicode.com/users/" + strconv.Itoa(id))
	if err != nil {
		panic(err)
	}

	return &UsersGetInstancesSession{
		NobodyRequestBuilder: gotcha.NobodyRequestBuilder{
			RequestMethod: http.MethodGet,
			RequestURL:    netURL,
		},
		JsonResponseHandler: gotcha.JsonResponseHandler{},
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
	agent := gotcha.NewAgent(http.DefaultClient)
	session := NewUsersGetInstancesSession(1)
	err := agent.Do(session)
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
