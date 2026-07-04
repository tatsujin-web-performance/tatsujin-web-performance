# 2章 モニタリング

## 2-7 実際にモニタリングを行う

### リスト1 CPU利用時間を表示するクエリ

```
avg without(cpu) (rate(node_cpu_seconds_total{mode!="idle"}[1m]))
```
