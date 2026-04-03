# houjin-cli 設計書

## 概要

法人番号システムWeb-API（国税庁）を操作するGo製CLIツール。法人番号（13桁）による法人情報の検索・取得を行う。

## 対象API

- **提供元**: 国税庁（法人番号公表サイト）
- **Base URL**: `https://api.houjin-bangou.nta.go.jp/4/`
- **APIバージョン**: v4固定
- **認証**: アプリケーションID（クエリパラメータ `id`）
- **レスポンス形式**: XML（`type=12`）— JSON非対応のため内部で変換
- **OpenAPI仕様**: なし（PDF仕様書のみ）

### エンドポイント

| パス | 機能 | CLIコマンド |
|------|------|------------|
| `/4/num` | 法人番号指定取得 | `houjin get` |
| `/4/name` | 法人名検索 | `houjin search` |
| `/4/diff` | 更新情報取得 | `houjin diff` |

## ディレクトリ構造

```
houjin-cli/
├── main.go                    # cobraルートコマンド実行
├── go.mod                     # github.com/planitaicojp/houjin-cli
├── CLAUDE.md
├── cmd/
│   ├── root.go                # グローバルフラグ: --format, --verbose, --config
│   ├── get.go                 # houjin get <法人番号...>
│   ├── search.go              # houjin search <法人名>
│   ├── diff.go                # houjin diff --from YYYY-MM-DD --to YYYY-MM-DD
│   ├── version.go             # houjin version
│   └── completion.go          # houjin completion bash/zsh/fish
├── internal/
│   ├── api/
│   │   ├── client.go          # HTTPクライアント（Base URL、App ID注入、XMLパース）
│   │   └── houjin.go          # GetByNumber, SearchByName, GetDiff
│   ├── config/
│   │   ├── config.go          # ~/.config/houjin/config.yaml ロード
│   │   └── env.go             # HOUJIN_APP_ID 環境変数オーバーライド
│   ├── model/
│   │   └── corporation.go     # Corporation構造体、チェックデジット検証
│   ├── output/
│   │   ├── formatter.go       # Formatterインターフェース
│   │   ├── json.go
│   │   ├── table.go
│   │   └── csv.go
│   └── errors/
│       └── errors.go          # APIError, ConfigError + 終了コード
└── testdata/                  # XMLレスポンスサンプル
```

**依存方向**: `cmd/ → internal/api/ → internal/model/`, `cmd/ → internal/output/`, `cmd/ → internal/config/`

## CLIコマンド詳細

### houjin get

法人番号を指定して法人情報を取得する。

```
houjin get <法人番号> [法人番号2 ...] [flags]
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--history` | bool | false | 履歴情報を含める |
| `--close` | bool | false | 閉鎖法人を含める |

- 複数法人番号を一度に照会可能（APIがセミコロン区切りで対応）
- 13桁チェックデジット事前検証、不正な番号はAPI呼び出し前にエラー返却

### houjin search

法人名で検索する。

```
houjin search <法人名> [flags]
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--mode` | string | prefix | 検索モード: `prefix`（前方一致）/ `partial`（部分一致） |
| `--pref` | string | - | 都道府県コード（01-47、99=海外） |
| `--city` | string | - | 市区町村コード |
| `--close` | bool | false | 閉鎖法人を含める |

- `--mode`はAPIの`mode=1/2`を英語キーワードにマッピング

### houjin diff

指定期間内の変更法人一覧を取得する。

```
houjin diff --from YYYY-MM-DD --to YYYY-MM-DD [flags]
```

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--from` | string | - | 開始日（必須） |
| `--to` | string | - | 終了日（必須） |

### グローバルフラグ

| フラグ | 型 | デフォルト | 説明 |
|--------|------|-----------|------|
| `--format` | string | json | 出力形式: `json` / `table` / `csv` |
| `--verbose` | bool | false | 詳細ログ出力 |
| `--config` | string | - | 設定ファイルパスのオーバーライド |

## APIクライアント

### client.go

- Base URL: `https://api.houjin-bangou.nta.go.jp/4/`
- 全リクエストに共通パラメータ `id`（App ID）と `type=12`（XML）を付与
- `encoding/xml`でXMLレスポンスをパース → `model.Corporation`に変換
- HTTPエラーハンドリング: 403=レートリミット、404=該当なし

### houjin.go

3つのメソッドを提供:

```go
func (c *Client) GetByNumber(numbers []string, opts GetOptions) (*model.Response, error)
func (c *Client) SearchByName(name string, opts SearchOptions) (*model.Response, error)
func (c *Client) GetDiff(from, to string) (*model.Response, error)
```

## データモデル

### Corporation

```go
type Corporation struct {
    CorporateNumber string  // 法人番号（13桁）
    Name            string  // 法人名
    NameKana        string  // 法人名カナ
    NameEnglish     string  // 英文法人名
    Kind            string  // 法人種別（101=国機関、201=地方公共団体、301=株式会社等）
    Prefecture      string  // 都道府県
    City            string  // 市区町村
    Address         string  // 詳細住所
    PostalCode      string  // 郵便番号
    AssignmentDate  string  // 法人番号指定日
    UpdateDate      string  // 最終更新日
    ChangeDate      string  // 変更日
    CloseDate       string  // 閉鎖日
    CloseCause      string  // 閉鎖事由
    Status          string  // ステータス
}
```

### Response

```go
type Response struct {
    Count        int
    DivideNumber int             // 分割番号（ページング）
    DivideSize   int             // 分割総数
    Corporations []Corporation
}
```

### チェックデジット検証

13桁法人番号のチェックデジット（先頭1桁）を算出・検証する関数を`model`パッケージに実装。`get`コマンドでAPI呼び出し前に事前検証を行い、不正な番号は即座にエラー返却。

## 設定管理

**優先順位**: コマンドフラグ > 環境変数 > 設定ファイル > デフォルト値

### 設定ファイル

```yaml
# ~/.config/houjin/config.yaml
app_id: "your-application-id"
format: json
```

### 環境変数

| 変数名 | 説明 |
|--------|------|
| `HOUJIN_APP_ID` | アプリケーションID（設定ファイルより優先） |

設定ファイルも環境変数もない場合、エラーメッセージとともに設定方法を案内する。

## 出力フォーマット

```go
type Formatter interface {
    Format(w io.Writer, resp *model.Response) error
}
```

| 形式 | 実装 | 説明 |
|------|------|------|
| JSON | `encoding/json` + indent | デフォルト、AIエージェント連携向け |
| Table | `text/tabwriter` | 人間が読みやすい固定幅テーブル |
| CSV | `encoding/csv` | ヘッダー行付き |

外部依存なし、Go標準ライブラリのみで実装。

## エラー処理

| 終了コード | 意味 |
|-----------|------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | 設定エラー（App ID未設定等） |
| 3 | APIエラー（403、サーバーエラー等） |
| 4 | 入力エラー（不正な法人番号等） |

## 依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `github.com/spf13/cobra` | CLIフレームワーク |
| `gopkg.in/yaml.v3` | 設定ファイルパース |
| その他すべて標準ライブラリ | `encoding/xml`, `encoding/json`, `encoding/csv`, `net/http`, `text/tabwriter` |

外部依存は2つのみ。

## テスト戦略

| レイヤー | テスト方式 |
|---------|-----------|
| `model/` | チェックデジット検証ユニットテスト |
| `output/` | 各フォーマッターの出力検証 |
| `api/` | `httptest.Server`でXMLレスポンスモック |
| `config/` | 環境変数/ファイル優先順位テスト |
| `cmd/` | 統合テスト（E2Eは後順位） |

`testdata/`ディレクトリに実際のAPIレスポンスXMLサンプルを保管し、パーステストに活用。

## 参考

- conoha-cli (`~/dev/crowdy/conoha-cli`) の構造を参照
- [法人番号システム Web-API](https://www.houjin-bangou.nta.go.jp/webapi/index.html)
- [WEB-API仕様書](https://www.houjin-bangou.nta.go.jp/webapi/kyuusiyousyo.html)
