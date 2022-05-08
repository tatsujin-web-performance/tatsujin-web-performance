package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

func main() {
	var err error
	// データベースへの接続
	db, err = sqlx.Open("mysql", "isuconp:isuconp@tcp(127.0.0.1:3306)/isuconp?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	results := []Post{}
	// 投稿一覧の取得
	err = db.Select(&results, "SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` ORDER BY `created_at` DESC LIMIT 30")
	if err != nil {
		log.Fatal(err)
	}
	// ユーザーIDリストを作る ❶
	userIDs := make([]int, 0)
	for _, p := range results {
		userIDs = append(userIDs, p.UserID)
	}
	// ユーザー情報をプリロードする ❷
	users := preloadUsers(userIDs)
	for _, p := range results {
		p.User = users[p.UserID]
	}
	out, _ := json.Marshal(results)
	fmt.Fprint(os.Stdout, string(out))
}

// データベースからユーザー情報を一括で取得する関数
func preloadUsers(ids []int) map[int]User {
	// 結果をいれるmap(連想配列)を作成。キーはユーザーID
	users := map[int]User{}
	// ユーザーリストが空の場合
	if len(ids) == 0 {
		return users
	}
	// ユーザーID用のリスト
	params := make([]interface{}, 0)
	// プレースホルダ用のリスト
	placeholders := make([]string, 0)
	for _, id := range ids {
		params = append(params, id)
		// プレースホルダ用のリストには'?'を入れる
		placeholders = append(placeholders, "?")
	}
	us := []User{}
	// IN句を利用してデータベースからユーザー情報を取得 ❸
	// プレースホルダのリストは','で連結してクエリを作成する
	err := db.Select(
		&us,
		"SELECT * FROM `users` WHERE `id` IN ("+strings.Join(placeholders, ",")+")",
		params...,
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range us {
		users[u.ID] = u
	}
	return users
}

// list-15 N+1データベースのプリロード(sqlx.Inを使う方法)
func preloadUsersIn(ids []int) map[int]User {
	// 結果をいれるmap(連想配列)を作成。キーはユーザーID
	users := map[int]User{}
	// ユーザーリストが空の場合
	if len(ids) == 0 {
		return users
	}
	// IN句を含むクエリを構築
	// query: プレースホルダ展開されたクエリ
	// params: クエリ実行時に渡すパラメータ
	query, params, err := sqlx.In(
		"SELECT * FROM `users` WHERE `id` IN (?)",
		ids,
	)
	if err != nil {
		log.Fatal(err)
	}
	us := []User{}
	// データベースからユーザー情報を取得
	err = db.Select(
		&us,
		query,
		params...,
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range us {
		users[u.ID] = u
	}
	return users
}
