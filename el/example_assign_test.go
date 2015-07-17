package el_test

import (
	"encoding/json"
	"fmt"

	"github.com/pascaldekloe/goe/el"
)

type Feed struct {
	About map[string]*Channel `json:"channels"`
}

type Channel struct {
	Items []*Item `json:"items"`
}

type Item struct {
	Title     string `json:"title"`
	Timestamp int64  `json:"created"`
}

func ExampleAssign() {
	f := new(Feed)
	el.Assign(f, `/About["demo"]/Items[1]/Title`, "Second")
	el.Assign(f, `/About[*]/Items[*]/Timestamp`, 1437146613)

	bytes, err := json.Marshal(f)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(bytes))
	// Output: {"channels":{"demo":{"items":[{"title":"","created":1437146613},{"title":"Second","created":1437146613}]}}}
}
