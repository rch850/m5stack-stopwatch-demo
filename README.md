# M5Stack StopWatch demos

M5Stack StopWatch 向けの TinyGo デモ集です。各デモのディレクトリでビルド・書き込みを実行します。

| ディレクトリ | デモ |
| --- | --- |
| [crab/](crab/README.md) | 左右のボタンでハサミを開閉し、マイクの音量に応じて口を動かすカニ |

```sh
cd crab
go test ./...
tinygo flash -target=./targets/stopwatch.json -port=/dev/cu.usbmodem101 .
```

書き込み手順とプレビューは各デモの README を参照してください。
