package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// UNIX time で疑似乱数のシード値を初期化
	rand.Seed(time.Now().Unix())

	// Users を新しく生成
	users := NewUsers(
		// まず連番 ID の5名を生成し
		WithSequentialIDUsers(5),
		// 次にランダム ID の5名を生成をし
		WithRandomIDUsers(5),
		// 最後にまた連番 ID の5名を生成
		WithSequentialIDUsers(5),
	)

	// 登録されいてる User 数を標準出力に表示
	fmt.Printf("users count: %d\n", users.Len())

	// 登録されいてる User の ID を登録順に順次標準出力に表示
	users.ForEach(func(i int, u *User) {
		fmt.Printf("%02d: user id: %d\n", i+1, u.ID)
	})
}

// User は ID だけを持つシンプルな構造体
type User struct {
	ID int
}

// User のリストと ID マップを持ち、読み書きロックを持つ構造体
type Users struct {
	mu   sync.RWMutex
	list []*User
	dict map[int]*User
}

// Users 生成時に指定できる UsersOption の型定義
// このように関数としてオプションを取れるようにするのが Functional Option パターン
type UsersOption func(users *Users)

// Users を生成する関数
// 可変長引数で UsersOption を受け取ることで、複数の UsersOption を受け取れるように作る
func NewUsers(opts ...UsersOption) *Users {
	// Users の生成
	users := &Users{
		// sync.RWMutex は値型のフィールドなので生成コードを書かなくても問題ない
		// mu:   sync.RWMutex{},

		// list はスライス、dict は map なので初期化時に生成しておかないと
		// nil になってしまう
		list: make([]*User, 0),
		dict: make(map[int]*User, 0),
	}

	// 引数に渡された UsersOption を順次実行
	for _, opt := range opts {
		opt(users)
	}

	// すべての UsersOption を適用した Users を返す
	return users
}

// Users に User を追加するメソッド
// ID がすでに登録済みなら false を返し登録しない
func (u *Users) Add(user *User) (ok bool) {
	// 書き込みのロックをとり、関数から出るときにロックを解除
	u.mu.Lock()
	defer u.mu.Unlock()

	// 追加しようとした User の ID が0以下なら追加しない
	if user.ID <= 0 {
		return false
	}

	// 追加しようとした User の ID がすでに登録済みなら追加しない
	if _, found := u.dict[user.ID]; found {
		return false
	}

	u.list = append(u.list, user)
	u.dict[user.ID] = user
	return true
}

// Users に登録されている User の数を返すメソッド
func (u *Users) Len() int {
	// 読み込みのロックを取り、関数から出る時にロックを解除
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Users.list の len は登録済みの User の数
	return len(u.list)
}

// Users に登録されている User 全てに対して追加順に関数を実行するメソッド
func (u *Users) ForEach(f func(i int, u *User)) {
	// 読み込みのロックを取り、関数から出る時にロックを解除
	u.mu.RLock()
	defer u.mu.RUnlock()

	// リストをループで回して順番に関数を実行
	for i, u := range u.list {
		f(i, u)
	}
}

// 登録されている User の持つ ID の中で最大の ID を返すメソッド
func (u *Users) MaxID() int {
	// 読み込みのロックを取り、関数から出る時にロックを解除
	u.mu.RLock()
	defer u.mu.RUnlock()

	// User が1人も登録されていなければ0
	maxID := 0
	for _, user := range u.list {
		// 現在の maxID より User の ID が大きければ上書き
		if user.ID > maxID {
			maxID = user.ID
		}
	}

	return maxID
}

// Users に count 引数の数だけ連番 ID の User を追加するオプション
func WithSequentialIDUsers(count int) UsersOption {
	return func(u *Users) {
		for i := 0; i < count; i++ {
			// 登録済みの最大 ID を取得し、それに1足した ID で採番
			id := u.MaxID() + 1
			user := &User{
				ID: id,
			}

			// 最大 ID より大きいので必ず追加に成功するものとして追加の成功はチェックしない
			u.Add(user)
		}
	}
}

// Users に count 引数の数だけランダムな ID の User を追加するオプション
func WithRandomIDUsers(count int) UsersOption {
	return func(u *Users) {
		for i := 1; i <= count; i++ {
			// 登録済みの最大 ID に +50 から -50 の範囲でランダムに生成した ID で採番
			id := u.MaxID() + rand.Intn(101) - 50
			user := &User{
				ID: id,
			}

			// ID が重複したり、マイナスになってしまって登録できず失敗する可能性があるので
			// 失敗した場合は要求された数だけ生成できるようにループカウンタを1つ戻す
			ok := u.Add(user)
			if !ok {
				i--
			}
		}
	}
}