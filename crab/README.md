# StopWatch Crab 🦀

M5Stack **StopWatch (C152 / ESP32-S3)** にカニを表示する TinyGo アプリです。

以下のコマンドは、この `crab/` ディレクトリで実行してください。

- 黄色の A ボタン（GPIO2）を押すと画面左のハサミが閉じ、離すと開きます。
- 青色の B ボタン（GPIO1）を押すと画面右のハサミが閉じ、離すと開きます。
- 同時押しにも対応。4 段階で動き、ボタンのチャタリングを 15 ms で除去します。

![カニの動作プレビュー](docs/crab.gif)

[実機での動作動画を見る（X）](https://x.com/rch850/status/2118732882000986221)

## ビルド

[TinyGo](https://tinygo.org/getting-started/install/) と Go をインストールして実行します。
TinyGo **0.44.0** / Go **1.27.3** / macOS arm64 でビルドを確認しました。
外部 Go モジュールへの依存はありません。

```sh
go test ./...
mkdir -p build
tinygo build -target=./targets/stopwatch.json -o build/crab.bin .
```

生成物は `build/crab.bin` です。`targets/stopwatch.json` が ESP32-S3 の汎用ターゲットを継承します。
`-target=m5stack` は別の ESP32 機種用なので使いません。

`make` を使える環境では `make test` / `make build` でも実行できます。コンパイラのパスを指定する場合:

```sh
make build TINYGO=/path/to/tinygo
```

## USB で書き込む

**書き込みを実行すると、現在のファームウェアが置き換わります。**
必要なら先に手元のファームウェアや設定をバックアップしてください。

1. StopWatch をデータ通信対応 USB-C ケーブルで接続します。
2. 電源ボタンを緑 LED が点滅するまで長押しし、離してダウンロードモードにします。
3. macOS で `ls /dev/cu.usbmodem*` を実行してポートを確認します。
4. 次を実行します（ポート名は実際のものに置き換えます）。

```sh
tinygo flash -target=./targets/stopwatch.json -port=/dev/cu.usbmodem101 .
```

`make flash PORT=/dev/cu.usbmodem101` も同じ操作です。
TinyGo 内蔵の ESP32 書き込み機能を使います。終了後に画面が出なければ電源ボタンを一度押して再起動してください。
初期化に失敗した場合、USB シリアルへエラーを繰り返し出力します。
エラーには失敗した I²C アドレス・レジスタ番号も表示されます。
書き込み後に画面が暗い場合は、電源ボタンを一度短く押して再起動してください。
実機調査では、初回に `i2c: error: expected ACK not NACK` で電源初期化が停止した後、
エラー詳細を追加した版の再書き込みと再起動で表示・操作が成功しました。
I²C エラーが発生した根本原因と、再現条件は未特定です。
元の工場ファームウェアへ戻す方法は [M5Stack の公式ガイド](https://docs.m5stack.com/en/guide/stopwatch/usage) を参照してください。

## PC で見た目を確認する

```sh
go run ./cmd/preview
```

実機と共通の描画コードから `docs/crab.png` と `docs/crab.gif` を生成します。
GIF は左・右・両方の順にハサミを動かすデモで、実機のフレームレートを示すものではありません。

## 構成と実機確認の状況

| ファイル | 役割 |
| --- | --- |
| `main.go` | GPIO 入力、左右独立の開閉、変更部分の更新 |
| `internal/crab/` | RGB565 のカニ描画、デバウンス、開閉状態 |
| `internal/stopwatch/power.go` | M5PM1 / M5IOE1 電源・リセットシーケンス |
| `internal/stopwatch/display.go` | CO5300 の初期化とソフトウェア QSPI |
| `cmd/preview/` | 同じ描画コードによる PNG / GIF 出力 |

画面の公称解像度 466×466 に描画します。M5GFX の StopWatch 設定に合わせてコントローラの X オフセットを +6 としています
（公式ドライバの論理サイズは 468×468）。画像バッファは 1 行 932 バイトで、PSRAM は使用しません。
初回だけ全面を描き、その後はハサミと操作インジケーターの範囲だけを描き直します。

TinyGo の公開 SPI API に Quad 転送がないため、GPIO の volatile レジスタ書き込みで QSPI を実装しています。
描画中は入力処理が待つため、ごく短い押下は拾えない場合があります。
このプロジェクトでは **コンパイル、入力・描画範囲・I2C エラー処理のテスト、画像プレビューを確認済み** です。
**実機への書き込みと、ユーザーによるカニの表示・ボタン操作を確認済み** です。
冷間起動や USB を外したバッテリー動作の安定性は未確認です。

タッチ、音、振動、無線は使用していません。

## 参照した公式資料

- [StopWatch のピン配置・仕様・ダウンロードモード](https://docs.m5stack.com/en/core/StopWatch)
- [M5GFX: StopWatch パネル初期化](https://github.com/m5stack/M5GFX/blob/master/src/M5GFX.cpp)
- [M5GFX: PMIC 操作](https://github.com/m5stack/M5GFX/blob/master/src/board_detect/m5/pmic_ops.hpp)
- [M5GFX: AMOLED 通信形式](https://github.com/m5stack/M5GFX/blob/master/src/lgfx/v1/panel/Panel_AMOLED.inl)
- [TinyGo: ESP32-S3 汎用ターゲット](https://github.com/tinygo-org/tinygo/blob/v0.44.0/targets/esp32s3-generic.json)
