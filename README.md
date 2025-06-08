# gphoto-cli
Google Photos Picker APIを利用したCLI

## 機能
- 🖼️ Google Photos Picker APIで全ライブラリからの写真選択
- 💾 画像ダウンロード機能
- 🔍 詳細なEXIF情報表示
- ⚙️ 対話式セットアップ

## セットアップ

### 1. アプリケーションのビルド
```bash
go build -o gphoto-cli
```

### 2. 対話式セットアップ
```bash
./gphoto-cli setup
```

このコマンドで以下が実行されます：
1. Google Cloud Console でのセットアップ手順を案内
2. OAuth 2.0 クライアント ID とシークレットの入力
3. 認証方式の選択（自動/手動）
4. 設定ファイル（`~/.gphoto-cli/config.yaml`）への保存
5. 認証トークン（`~/.gphoto-cli/token.json`）の保存

### 3. 設定管理
```bash
# 現在の設定を確認
./gphoto-cli config show

# 設定をリセット
./gphoto-cli config reset
```

## 使用方法

### 基本的な写真選択とメタデータ表示
```bash
# Google Photos Pickerで写真を選択し、詳細情報を表示
./gphoto-cli picker
```

### 画像ダウンロード
```bash
# 指定ディレクトリにダウンロード
./gphoto-cli download --output ./my-photos

# サムネイルサイズでダウンロード
./gphoto-cli download --thumbnail

# デフォルト（~/gphoto-downloads）にダウンロード
./gphoto-cli download
```

### クイックビューモード
```bash
# 写真選択とメタデータ表示
./gphoto-cli view
```

### その他のコマンド
```bash
# バージョン表示
./gphoto-cli version

# ヘルプ表示
./gphoto-cli --help
```

## コマンド詳細

### picker
Google Photos Picker APIを使用してライブラリ全体から写真を選択し、以下の情報を表示します：
- ファイル名、ID、タイプ
- 作成日時、サイズ
- カメラ情報（メーカー、モデル）
- 撮影設定（絞り、焦点距離、ISO、シャッタースピード）
- BaseURL

### download
Google Photos Picker APIで選択した写真をローカルディレクトリにダウンロードします：
- `--output` (`-o`): 出力ディレクトリを指定（デフォルト: ~/gphoto-downloads）
- `--thumbnail`: サムネイルサイズでダウンロード（高速）

### view
pickerコマンドと同じ機能を提供するクイックビューモードです。