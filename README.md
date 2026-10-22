codemirror/collabの概念実証

```sh
# relay 側: 鍵を作る。公開鍵の1行が表示される
./myapp keygen

# メイン側: authorized_keys にその1行を追記して起動
echo 'ssh-ed25519 AAAA... ' >> authorized_keys
./myapp serve --addr 0.0.0.0:3002

# relay 側: 起動
./myapp relay --upstream ws://127.0.0.1:3002 --addr 127.0.0.1:3004
```
