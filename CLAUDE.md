# overview

CodeMirror6/Collabの概念実証。
y-codemirror, y-websocketを使ったリアルタイム共同編集基盤からcodemirror/collabを使った基盤に置き換えられないか検証する。

## 機能

- 現在、同じエディアを開いているユーザとカーソルを同期させる


## rules

- 後方互換性は維持しなくてよい。
- when fixing bugs, add a failing regression test first.
- all errors are user-facing, so messages should be clear.
- keep functions small and focused.
- module files should re-export what's needed, hide implementation details.
- 変更内容を codex形式(search/replace形式)で出力してください。
例）
```
mathweb/flask/app.py
<<<<<<< search
from flask import flask
=======
import math
from flask import flask
>>>>>>> replace
```
- ファイルを削除・移動するときはrm・mvコマンドで提示する。
- 全コードを書き直すときは、古いファイルをsearch/replaceせずに削除して、新規ファイルとして出力する。
例)
internal/parser/links_test.go (new file)
```go
package parser
import (
	"slices"
	"testing"
)
```

- jsxにおいて、return の先頭にコメント（{/*...*/} ）を置く場合は fragment （<>...</>）で囲まなければならない。
- 機械的な置換にはperl言語を使ってください

## tech stack
### backend
- go 1.26.0 (1.27はライブラリ非対応)
- coder/websocket v1.8

### frontend
- solid.js v1.9
- codemirror/collab (v6)

