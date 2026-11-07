# StopWatch Crab 🦀

M5Stack **StopWatch (C152 / ESP32-S3)** にカニを表示する TinyGo アプリです。

以下のコマンドは、この `crab/` ディレクトリで実行してください。

- 黄色の A ボタン（GPIO2）を押すと画面左のハサミが閉じ、離すと開きます。
- 青色の B ボタン（GPIO1）を押すと画面右のハサミが閉じ、離すと開きます。
- 同時押しにも対応。4 段階で動き、ボタンのチャタリングを 15 ms で除去します。
- 内蔵マイクに向かって話したり手を叩いたりすると、音量に応じて口が 8 段階で開きます。静かになると滑らかに閉じます。

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
初期化に失敗した場合、USB シリアルへエラーを出力し、500 ms 待って初期化を再試行します。
エラーには失敗した I²C アドレス・レジスタ番号も表示されます。
旧版では `power read addr=0x4f reg=0x23: i2c: error: expected ACK not NACK` で停止する事例がありました。
M5IOE1 の I²C 設定を初期化する前に読み出していたため、公式ドライバと同じく先に `0x23` へ直接書き込む順序に修正しました。
読み出しに加えて書き込みも一時的な NACK を再試行します。修正版の書き込み後、実機で正常起動と表示・操作・マイク反応を確認済みです。
書き込み直後は数秒待ち、それでも画面が暗い場合は電源ボタンを一度短く押してください。
改善しない場合は USB シリアルの起動ログに出るアドレス・レジスタ番号を確認してください。
元の工場ファームウェアへ戻す方法は [M5Stack の公式ガイド](https://docs.m5stack.com/en/guide/stopwatch/usage) を参照してください。

## PC で見た目を確認する

```sh
go run ./cmd/preview
```

実機と共通の描画コードから `docs/crab.png` と `docs/crab.gif` を生成します。
GIF は左・右・両方のハサミと口の開閉を順に動かすデモで、実機のフレームレートを示すものではありません。

## マイクの反応

起動後、マイクのウォームアップに約 1 秒かかります。16 kHz / 16 bit で取得した音声の
直流成分を除いた RMS（実効値）から口の開き具合を決めます。小さなノイズを無視し、
音が鳴るとすばやく開き、音が止まるとゆっくり閉じます。

反応の強さは `internal/crab/volume.go` の `NoiseGate`（小さい音を無視するしきい値）と
`FullVolume`（口が最大まで開く音量）で調整できます。`FullVolume` を下げると、小さな声でも大きく開きます。
値は符号付き 16 bit PCM の振幅で、校正した dB 値ではありません。

音声は口の動きの計算に使い、その場で破棄します。保存・ネットワーク送信・スピーカー再生はしません。
マイクの初期化に失敗しても、ハサミのボタン操作は継続します。USB シリアルにエラーの詳細を出力します。

## 構成と実機確認の状況

| ファイル | 役割 |
| --- | --- |
| `main.go` | GPIO 入力、マイクの取得、ハサミと口の変更部分の更新 |
| `internal/crab/` | RGB565 のカニ描画、デバウンス、音量から口の開き具合への変換 |
| `internal/stopwatch/power.go` | M5PM1 / M5IOE1 電源・リセットシーケンス |
| `internal/stopwatch/display.go` | CO5300 の初期化とソフトウェア QSPI |
| `internal/stopwatch/codec.go` / `microphone.go` | ES8311 の電源・ADC 初期化と I2S 音声取得 |
| `cmd/preview/` | 同じ描画コードによる PNG / GIF 出力 |

画面の公称解像度 466×466 に描画します。M5GFX の StopWatch 設定に合わせてコントローラの X オフセットを +6 としています
（公式ドライバの論理サイズは 468×468）。画像バッファは 1 行 932 バイトで、PSRAM は使用しません。
初回だけ全面を描き、その後はハサミ・操作インジケーター・口の変更範囲だけを描き直します。
マイクの ES8311 は BCLK をクロック源として使うため、TinyGo の MCLK 出力は不要です。

TinyGo の公開 SPI API に Quad 転送がないため、GPIO の volatile レジスタ書き込みで QSPI を実装しています。
描画中は入力処理が待つため、ごく短い押下は拾えない場合があります。
このプロジェクトでは **コンパイル、入力・音量変換・描画範囲・I2C エラー処理のテスト、画像プレビューを確認済み** です。
**実機への書き込みと、ユーザーによるカニの表示・ボタン操作を確認済み** です。
マイク対応版も書き込み済みで、USB シリアルの `Microphone capture active (16000 Hz)` まで確認しています。
ユーザーによる実機確認で、音に合わせて口が動くことも確認済みです。感度は利用環境に合わせて調整してください。
冷間起動や USB を外したバッテリー動作の安定性は未確認です。

タッチ、スピーカー、振動、無線は使用していません。

## 参照した公式資料

- [StopWatch のピン配置・仕様・ダウンロードモード](https://docs.m5stack.com/en/core/StopWatch)
- [M5GFX: StopWatch パネル初期化](https://github.com/m5stack/M5GFX/blob/master/src/M5GFX.cpp)
- [M5GFX: PMIC 操作](https://github.com/m5stack/M5GFX/blob/master/src/board_detect/m5/pmic_ops.hpp)
- [M5GFX: AMOLED 通信形式](https://github.com/m5stack/M5GFX/blob/master/src/lgfx/v1/panel/Panel_AMOLED.inl)
- [TinyGo: ESP32-S3 汎用ターゲット](https://github.com/tinygo-org/tinygo/blob/v0.44.0/targets/esp32s3-generic.json)
- [M5Unified: StopWatch のマイク設定とクロック開始後の ADC 有効化](https://github.com/m5stack/M5Unified/blob/master/src/M5Unified.inl)
