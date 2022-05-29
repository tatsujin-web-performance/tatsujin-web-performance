package main

import (
  "fmt"
)

// fmt.Stringer を満たす Dog 構造体
type Dog struct {
  Name string
  Age  int
}

// fmt.Stringer.String の実装
func (d *Dog) String() string {
  return fmt.Sprintf("Dog: %s (%d)", d.Name, d.Age)
}

// fmt.GoStringer は実装されていないので標準の挙動が使われる
// func (d *Dog) GoString() string {
//   // TODO: Implmenet fmt.GoStringer
// }

// fmt.Stringer と fmt.GoStringer を満たす Cat 構造体
type Cat struct {
  Name string
  Age  int
}

// fmt.Stringer.String の実装
func (c *Cat) String() string {
  return fmt.Sprintf("(=^_^=) Cat: %s(%d)", c.Name, c.Age)
}

// fmt.GoStringer.GoString の実装
func (c *Cat) GoString() string {
  return fmt.Sprintf("(=^_^=) &Cat{%#v, %#v}", c.Name, c.Age)
}

func main() {
  // Dog の生成
  dog := &Dog{Name:"coco", Age:5}

  // Cat の生成
  cat := &Cat{Name:"nana", Age:3}

  fmt.Println(dog) // => Dog: coco (5)
  fmt.Println(cat) // => (=^_^=) Cat: nana(3)

  fmt.Printf("%#v\n", dog) // => &main.Dog{Name:"coco", Age:5}
  fmt.Printf("%#v\n", cat) // => (=^_^=) &Cat{"nana", 3}
}
