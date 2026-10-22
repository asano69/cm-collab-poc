codemirror/collabの概念実証

```sh
# relay 側: 鍵を作る。公開鍵の1行が表示される
./myapp keygen

# メイン側: authorized_keys にその1行を追記して起動
echo 'ssh-ed25519 AAAA... ' >> authorized_keys
./myapp serve --addr 0.0.0.0:3002

# relay 側: メインが書き出した known_hosts をコピーして起動
cp <メインのディレクトリ>/known_hosts .
./myapp relay --upstream ws://127.0.0.1:3002 --addr 127.0.0.1:3004
```

![](./example.png)
