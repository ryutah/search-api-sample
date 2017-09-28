package sapi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type User struct {
	Name      string
	Comment   string
	Visits    float64
	LastVisit time.Time
	Birthday  time.Time
}

type UserIndex struct {
	ID        string `search:"-"`
	Name      string
	Comment   search.HTML
	Visits    float64
	LastVisit time.Time
	Birthday  time.Time
}

func PutSamples(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	users := make(map[*datastore.Key]User)
	for i := 0; i < 500; i++ {
		key := datastore.NewIncompleteKey(ctx, "User", nil)
		usr := User{
			Name:      fmt.Sprintf("Sample User%d", i+1),
			Comment:   fmt.Sprintf("<p>Sample Comment%d</p>", i+1),
			Visits:    float64(rand.Int63n(100)),
			LastVisit: time.Now(),
			Birthday:  time.Date(1990, time.January, rand.Intn(29), 0, 0, 0, 0, time.UTC),
		}
		newKey, err := datastore.Put(ctx, key, &usr)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		users[newKey] = usr
	}

	index, err := search.Open("users")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for key, usr := range users {
		usrIndex := &UserIndex{
			Name:      usr.Name,
			Comment:   search.HTML(usr.Comment),
			Visits:    usr.Visits,
			LastVisit: usr.LastVisit,
			Birthday:  usr.Birthday,
		}
		id := fmt.Sprint(key.IntID())
		if _, err := index.Put(ctx, id, usrIndex); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	fmt.Println("OK")
}

func Search(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	index, err := search.Open("users")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	userName := r.FormValue("name")
	cursor := r.FormValue("cursor")
	query := fmt.Sprintf("Name = %s", userName)
	log.Infof(ctx, "Query: %v", query)

	option := &search.SearchOptions{
		Sort: &search.SortOptions{
			Expressions: []search.SortExpression{
				{
					Expr:    "Birthday",
					Reverse: false,
				},
			},
		},
		Limit:  200,
		Cursor: search.Cursor(cursor),
	}

	ite := index.Search(ctx, query, option)

	type result struct {
		Users  []*UserIndex
		Cursor search.Cursor
	}
	rslt := new(result)
	for {
		usrIndex := new(UserIndex)
		id, err := ite.Next(usrIndex)
		if err == search.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		usrIndex.ID = id
		rslt.Users = append(rslt.Users, usrIndex)
	}
	log.Infof(ctx, "%v", ite.Cursor())
	rslt.Cursor = ite.Cursor()

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	if err := json.NewEncoder(w).Encode(rslt); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
