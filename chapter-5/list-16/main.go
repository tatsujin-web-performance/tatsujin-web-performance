package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// ユーザー情報の構造体 DBのカラム名、JSONでのキー名を付与している
type User struct {
	ID          int       `db:"id" json:"id"`
	AccountName string    `db:"account_name" json:"account_name"`
	Passhash    string    `db:"passhash" json:"passhash"`
	Authority   int       `db:"authority" json:"authority"`
	DelFlg      int       `db:"del_flg" json:"del_flg"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// 投稿情報の構造体 DBのカラム名、JSONでのキー名を付与している
type Post struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Body      string    `db:"body" json:"body"`
	Mime      string    `db:"mime" json:"mime"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	User      User      `db:"user" json:"users"`
}

var db *sqlx.DB

func main() {
	var err error
	db, err = sqlx.Open("mysql", "isuconp:isuconp@tcp(127.0.0.1:3306)/isuconp?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	results := []Post{}
	// JOIN (INNER JOIN)による投稿一覧・ユーザー情報の取得 ❶
	query := "SELECT " +
		"p.id AS `id`, " +
		"p.user_id AS `user_id`," +
		"p.body AS `body`, " +
		"p.mime AS `mime`, " +
		"p.created_at AS `created_at`, " +
		"u.id AS `user.id`, " + // ユーザー情報のカラム名指定 ❷
		"u.account_name AS `user.account_name`, " +
		"u.passhash AS `user.passhash`, " +
		"u.authority AS `user.authority`, " +
		"u.del_flg AS `user.del_flg`," +
		"u.created_at AS `user.created_at` " +
		"FROM `posts` p JOIN `users` u ON p.user_id = u.id ORDER BY p.created_at DESC LIMIT 30"
	err = db.Select(&results, query)
	if err != nil {
		log.Fatal(err)
	}
	out, _ := json.Marshal(results)
	fmt.Fprint(os.Stdout, string(out))
}
