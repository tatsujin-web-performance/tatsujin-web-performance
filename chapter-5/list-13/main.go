package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID          int       `db:"id" json:"id"`
	AccountName string    `db:"account_name" json:"account_name"`
	Passhash    string    `db:"passhash" json:"passhash"`
	Authority   int       `db:"authority" json:"authority"`
	DelFlg      int       `db:"del_flg" json:"del_flg"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Post struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Body      string    `db:"body" json:"body"`
	Mime      string    `db:"mime" json:"mime"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	User      User      `json:"users"`
}

var db *sqlx.DB
var mc *memcache.Client

func main() {
	var err error
	// データベースへの接続
	db, err = sqlx.Open("mysql", "isuconp:@tcp(127.0.0.1:3306)/isuconp?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	// memcachedへの接続
	mc = memcache.New("127.0.0.1:11211")

	results := []Post{}
	err = db.Select(&results, "SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` ORDER BY `created_at` DESC LIMIT 30")
	if err != nil {
		log.Fatal(err)
	}
	// キャッシュから取得するユーザーIDリストを作る ❶
	userIDs := make([]int, 0)
	for _, p := range results {
		userIDs = append(userIDs, p.UserID)
	}
	// キャッシュからユーザー情報を一括で取得 ❷
	users := getUsers(userIDs)
	for _, p := range results {
		if u, ok := users[p.UserID]; ok {
			p.User = u
		} else {
			// キャッシュから取得できなかった場合、データベースから取得 ❹
			p.User = getUser(p.UserID)
		}
	}
	out, _ := json.Marshal(results)
	fmt.Fprint(os.Stdout, string(out))
}

// キャッシュからユーザー情報を一括で取得する関数
func getUsers(ids []int) map[int]User {
	// キャッシュのキーのリストを作成
	keys := make([]string, 0)
	for _, id := range ids {
		keys = append(keys, fmt.Sprintf("user_id:%d", id))
	}
	// 結果をいれるmap(連想配列)を作成。キーはユーザーID
	users := map[int]User{}
	// キャッシュから複数のキャッシュを取得 ❸
	items, err := mc.GetMulti(keys)
	if err != nil {
		return users
	}
	for _, it := range items {
		u := User{}
		// JSONをデコードし、mapにユーザーIDをキーとして格納
		err := json.Unmarshal(it.Value, &u)
		if err != nil {
			log.Fatal(err)
		}
		users[u.ID] = u
	}
	return users
}

func getUser(id int) User {
	user := User{}
	// データベースからユーザー情報を取得
	err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	// JSONにエンコード
	j, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	// キャッシュに格納 ❺
	mc.Set(&memcache.Item{
		Key:        fmt.Sprintf("user_id:%d", id),
		Value:      j,
		Expiration: 3600,
	})
	return user
}
