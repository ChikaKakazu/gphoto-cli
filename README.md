# gphoto-cli
Google Photos Picker APIを利用した画像表示対応CLI

## 機能
- 🖼️ Google Photos Picker APIで全ライブラリからの写真選択
- 📱 ターミナル内ASCII プレビュー
- 🖥️ 外部ビューアーでの画像表示
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

### 3. 設定管理
```bash
# 現在の設定を確認
./gphoto-cli config show

# 設定をリセット
./gphoto-cli config reset
```

## 使用方法

### 基本的な写真選択
```bash
./gphoto-cli picker
```

### 画像表示オプション
```bash
# ターミナル内プレビュー表示
./gphoto-cli picker --preview

# 外部ビューアーで開く
./gphoto-cli picker --open

# 画像をダウンロード（一時ディレクトリ）
./gphoto-cli picker --download

# サムネイルサイズでダウンロード
./gphoto-cli picker --thumbnail

# クイックビューモード（プレビュー + 外部表示）
./gphoto-cli view
```

### 専用ダウンロードコマンド
```bash
# 指定ディレクトリにダウンロード
./gphoto-cli download --output ./my-photos

# サムネイルサイズでダウンロード
./gphoto-cli download --thumbnail

# デフォルト（./downloads）にダウンロード
./gphoto-cli download
```

### その他のコマンド
```bash
# バージョン表示
./gphoto-cli version

# ヘルプ表示
./gphoto-cli --help
```