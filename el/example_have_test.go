package el_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pascaldekloe/goe/el"
)

type Feed struct {
	About map[string]*Channel `json:"channels"`
}

type Channel struct {
	Timestamp time.Time `json:"created"`
	Items     []*Item   `json:"items"`
}

type Item struct {
	Title *string `json:"title"`
}

func ExampleHave() {
	t, err := time.ParseInLocation(time.RFC3339, "2015-06-26T20:14:30+02:00", nil)
	if err != nil {
		panic(err.Error())
	}

	f := new(Feed)
	el.Have(f, `/About["demo"]/Items[1]/Title`, "Second")
	el.Have(f, `/About[*]/Timestamp`, t)

	bytes, err := json.Marshal(f)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(bytes))
	// Output: {"channels":{"demo":{"created":"2015-06-26T20:14:30+02:00","items":[null,{"title":"Second"}]}}}
}
