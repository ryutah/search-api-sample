package sapi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type User struct {
	ID        int64     `datastore:"-" json:"id"`
	Name      string    `json:"name"`
	Comment   string    `json:"comment"`
	Visits    float64   `json:"visits"`
	LastVisit time.Time `json:"lastVisit"`
	Birthday  time.Time `json:"birthday"`
	Mail      []string  `json:"mail"`
	UserID    int64     `json:"userId"`
	Field1    string    `json:"field1"`
	Field2    string    `json:"field2"`
}

type UserIndex struct {
	ID        string `search:"-"`
	Name      string
	Comment   search.HTML
	Visits    float64
	LastVisit time.Time
	Birthday  time.Time
	Mail      string
	UserID    string
	Field1    search.Atom `search:"Search"`
	Field2    search.Atom `search:"Search"`
}

func PutSamples(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	mails := [][]string{
		{"mail1@sample.com", "mail2@sample.com"},
		{"mail2@sample.com", "mail3@sample.com"},
		{"mail4@sample.com", "mail5@sample.com"},
	}

	users := make(map[*datastore.Key]User)
	for i := 0; i < 500; i++ {
		key := datastore.NewIncompleteKey(ctx, "User", nil)
		usr := User{
			Name:      fmt.Sprintf("Sample User%d", i+1),
			Comment:   fmt.Sprintf("<p>Sample Comment%d</p>", i+1),
			Visits:    float64(rand.Int63n(100)),
			LastVisit: time.Now(),
			Birthday:  time.Date(1990, time.January, rand.Intn(29), 0, 0, 0, 0, time.UTC),
			Mail:      mails[i%3],
			UserID:    int64(i + 1),
			Field1:    fmt.Sprintf("HOGE%v", i+1),
			Field2:    fmt.Sprintf("FUGA%v", i+1),
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
			Mail:      strings.Join(usr.Mail, " "),
			UserID:    fmt.Sprint(usr.UserID),
			Field1:    search.Atom(usr.Field1),
			Field2:    search.Atom(usr.Field2),
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

	mail := r.FormValue("mail")
	userID := r.FormValue("userid")
	cursor := r.FormValue("cursor")

	var query string
	if mail != "" {
		query += fmt.Sprintf(`Mail = "%s"`, mail)
	}
	if userID != "" {
		query += fmt.Sprintf(`UserID = "%s"`, userID)
	}
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

	var ids []int64
	for {
		id, err := ite.Next(nil)
		if err == search.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		i, _ := strconv.ParseInt(id, 10, 64)
		ids = append(ids, i)
	}

	var keys []*datastore.Key
	for _, id := range ids {
		key := datastore.NewKey(ctx, "User", "", id, nil)
		keys = append(keys, key)
	}
	users := make([]User, len(keys))
	if err := datastore.GetMulti(ctx, keys, users); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type result struct {
		Users  []User
		Cursor search.Cursor
	}
	rslt := result{
		Users:  users,
		Cursor: ite.Cursor(),
	}
	rslt.Cursor = ite.Cursor()

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	if err := json.NewEncoder(w).Encode(rslt); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func PostUser(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	payload := new(User)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var id string
	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		key := datastore.NewIncompleteKey(tc, "User", nil)
		newKey, err := datastore.Put(tc, key, payload)
		if err != nil {
			return err
		}
		id = fmt.Sprint(newKey.IntID())

		usrIndex := &UserIndex{
			Name:      payload.Name,
			Comment:   search.HTML(payload.Comment),
			Visits:    payload.Visits,
			LastVisit: payload.LastVisit,
			Birthday:  payload.Birthday,
			Mail:      strings.Join(payload.Mail, " "),
			UserID:    fmt.Sprint(payload.UserID),
			Field1:    search.Atom(payload.Field1),
			Field2:    search.Atom(payload.Field2),
		}

		index, err := search.Open("users")
		if err != nil {
			return err
		}
		_, err = index.Put(tc, id, usrIndex)
		return err
	}, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	result := map[string]string{"id": id}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
