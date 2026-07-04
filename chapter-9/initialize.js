// k6のhttp処理のmoduleをimport
import http from "k6/http";

// k6のsleep関数をimport
import { sleep } from "k6";

// 独自に定義したurl関数をimport
import { url } from "./config.js";

// k6が実行する関数
// /initializeに10秒のタイムアウトを指定してGETリクエストし、完了後1秒待機する
export default function () {
  http.get(url("/initialize"), {
    timeout: "10s",
  });
  sleep(1);
}
