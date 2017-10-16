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

type UsersGetInstancesRequester struct {
	gotcha.NobodyRequestBuilder
	gotcha.JsonResponseHandler
}

func NewUsersGetInstancesRequester(id int) *UsersGetInstancesRequester {
	netURL, err := url.Parse("https://jsonplaceholder.typicode.com/users/" + strconv.Itoa(id))
	if err != nil {
		panic(err)
	}

	return &UsersGetInstancesRequester{
		NobodyRequestBuilder: gotcha.NobodyRequestBuilder{
			RequestMethod: http.MethodGet,
			RequestURL:    netURL,
		},
		JsonResponseHandler: gotcha.JsonResponseHandler{},
	}
}

func (r *UsersGetInstancesRequester) ParseBody() (*User, error) {
	if !r.IsJSON() {
		return nil, fmt.Errorf("Response is not json: %s", string(r.GetBody()))
	}

	decoder := r.GetDecoder()
	decoder.UseNumber()

	var body User
	err := decoder.Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, err
}

func main() {
	agent := gotcha.NewAgent(http.DefaultClient)
	requester := NewUsersGetInstancesRequester(1)
	err := agent.Do(requester)
	if err != nil {
		log.Fatal(err)
	}

	res, err := requester.ParseBody()
	if err != nil {
		log.Fatal(err)
	}

	pp.Print(res)
}
```
