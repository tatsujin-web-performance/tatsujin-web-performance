// k6mからhttp処理のmoduleをimport
import http from "k6/http";

// k6からcheck関数をimport
import { check } from "k6";

// k6からHTMLをパースする関数をimport
import { parseHTML } from "k6/html";

// url関数をimport
import { url } from "./config.js";

// ベンチマーカーがが実行するシナリオ関数
// ログインしてからコメントを投稿する
export default function () {
  // /login に対してアカウント名とパスワードを送信
  const login_res = http.post(url("/login"), {
    account_name: "terra",
    password: "terraterra",
  });

  // レスポンスのステータスコードが 200 であることを確認
  check(login_res, {
    "is status 200": (r) => r.status === 200,
  });

  // ユーザーページ /@terra をGET
  const res = http.get(url("/@terra"));

  // レスポンスの内容をHTMLとして解釈
  const doc = parseHTML(res.body);

  // フォームのhidden要素から csrf_token, post_id を抽出
  const token = doc.find('input[name="csrf_token"]').first().attr("value");
  const post_id = doc.find('input[name="post_id"]').first().attr("value");

  // /comment に対して、post_id, csrf_token とともにコメント本文をPOST
  const comment_res = http.post(url("/comment"), {
    post_id: post_id,
    csrf_token: token,
    comment: "Hello k6!",
  });
  check(comment_res, {
    "is status 200": (r) => r.status === 200,
  });
}
