// SharedArray をimport
import { SharedArray } from "k6/data";

// accounts.json を読み込んで SharedArray にする
const accounts = new SharedArray("accounts", function () {
  return JSON.parse(open("./accounts.json"));
});

// SharedArray からランダムに1件取り出して返却する関数
export function getAccount() {
  return accounts[Math.floor(Math.random() * accounts.length)];
}
