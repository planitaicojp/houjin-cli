# houjin-cli

国税庁 法人番号システム Web-API を操作する CLI ツール。

## 特徴

- シングルバイナリ、クロスプラットフォーム（Linux / macOS / Windows）
- 法人番号による法人情報取得、法人名検索、更新情報取得
- 13桁チェックデジット事前検証
- 構造化出力（JSON / テーブル / CSV）
- AIエージェント連携に最適化（JSON デフォルト出力）

## インストール

### Homebrew（macOS / Linux）

```bash
brew install planitaicojp/tap/houjin
```

### Scoop（Windows）

```powershell
scoop bucket add planitaicojp https://github.com/planitaicojp/bucket
scoop install houjin
```

### GitHub Releases

[Releases](https://github.com/planitaicojp/houjin-cli/releases) ページからバイナリをダウンロード。

### ソースからビルド

```bash
go install github.com/planitaicojp/houjin-cli@latest
```

## セットアップ

[法人番号システム Web-API](https://www.houjin-bangou.nta.go.jp/webapi/) でアプリケーション ID を取得し、以下のいずれかで設定：

```bash
# 環境変数
export HOUJIN_APP_ID=your-application-id

# または設定ファイル
mkdir -p ~/.config/houjin
cat > ~/.config/houjin/config.yaml << EOF
app_id: your-application-id
format: json
EOF
```

優先順位: コマンドフラグ > 環境変数 > 設定ファイル

## 使い方

### 法人番号で取得

```bash
# 法人情報を取得
houjin get 1180301018771

# 複数番号を一度に取得
houjin get 1180301018771 5180301018778

# 履歴情報を含める
houjin get 1180301018771 --history

# テーブル形式で出力
houjin get 1180301018771 --format table
```

### 法人名で検索

```bash
# 前方一致検索（デフォルト）
houjin search トヨタ

# 部分一致検索
houjin search トヨタ --mode partial

# 都道府県を指定して検索
houjin search トヨタ --pref 23
```

### 更新情報の取得

```bash
# 期間内に更新された法人一覧
houjin diff --from 2024-01-01 --to 2024-01-31
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `houjin get <番号...>` | 法人番号を指定して法人情報を取得 |
| `houjin search <法人名>` | 法人名で検索 |
| `houjin diff` | 指定期間内の更新法人一覧を取得 |
| `houjin version` | バージョン情報を表示 |
| `houjin completion` | シェル補完スクリプトを生成 |

## グローバルフラグ

| フラグ | 説明 |
|--------|------|
| `--format` | 出力形式: `json`（デフォルト）, `table`, `csv` |
| `--verbose` | 詳細ログ出力 |
| `--config` | 設定ファイルパスの指定 |

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | 設定エラー（アプリケーション ID 未設定等） |
| 3 | API エラー（レートリミット、サーバーエラー等） |
| 4 | 入力エラー（不正な法人番号等） |

## エージェント連携

```bash
# 非対話モード + JSON出力（デフォルト）で利用
HOUJIN_APP_ID=your-id houjin get 1180301018771

# 終了コードで分岐
if houjin get "$CORP_NUM" > /tmp/result.json 2>/dev/null; then
  cat /tmp/result.json | jq '.corporations[0].name'
else
  echo "取得失敗 (exit: $?)"
fi
```

## 開発

```bash
make build     # ビルド
make test      # テスト実行
make lint      # lint実行
make clean     # クリーンアップ
make coverage  # カバレッジレポート生成
```

## API ドキュメント

- [法人番号システム Web-API](https://www.houjin-bangou.nta.go.jp/webapi/)
- [Web-API 仕様書](https://www.houjin-bangou.nta.go.jp/webapi/kyuusiyousyo.html)

## ライセンス

MIT
