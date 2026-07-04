// 各ファイルからシナリオ関数を import
import initialize from "./initialize.js";
import comment from "./comment.js";
import postimage from "./postimage.js";

// k6が各関数を実行できるようにexport
export { initialize, comment, postimage };

// 複数のシナリオを組み合わせて実行するオプションの定義
export const options = {
  scenarios: {
    initialize: {
      executor: "shared-iterations", // 一定量の実行を複数のVUsで共有する実行機構
      vus: 1, // 同時実行数(初期化なので1)
      iterations: 1, // 繰返し回数(初期化なので1回だけ)
      exec: "initialize", // 実行するシナリオの関数名
      maxDuration: "10s", // 最大実行時間
    },
    comment: {
      executor: "constant-vus", // 複数の VUs を並行で動かす実行機構
      vus: 4, // 4 VUs で実行
      duration: "30s", // 30秒間実行する
      exec: "comment", // comment 関数を実行
      startTime: "12s", // 12秒後に実行開始
    },
    postImage: {
      executor: "constant-vus",
      vus: 2,
      duration: "30s",
      exec: "postimage",
      startTime: "12s",
    },
  },
};

// k6が実行する関数。定義は空でよい
export default function () {}
