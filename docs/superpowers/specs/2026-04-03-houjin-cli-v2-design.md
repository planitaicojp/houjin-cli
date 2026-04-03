# houjin-cli v0.2.0 設計書

## 概要

houjin-cli の v0.1.0 リリースおよび追加機能（ページング、法人種別フィルタ、変更事由フィルタ、住所パラメータ拡張、バッチ入力）と App ID 申請ガイドの設計。

---

## Phase 0: v0.1.0 リリーステスト

### 目的

goreleaser による初回リリースの動作確認。

### 変更点

#### `.goreleaser.yaml`

- `brews` セクションに `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"` を追加
- `scoops` セクションに `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"` を追加
- `scoops` の repository name を `scoop-bucket` から `bucket` に変更（freee-cli と統一）

#### `.github/workflows/release.yml`

- goreleaser ステップに環境変数を追加:
  - `HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}`

#### リリース手順

1. 上記修正を PR でマージ
2. `git tag v0.1.0` → `git push origin v0.1.0`
3. GitHub Actions release workflow が goreleaser 実行
4. 確認項目:
   - GitHub Releases にバイナリ（linux/darwin/windows × amd64/arm64）生成
   - checksums.txt 生成
   - `planitaicojp/homebrew-tap` に formula 生成
   - `planitaicojp/bucket` に manifest 生成

---

## Phase 1: 追加機能

### 1a) ページング対応

#### 背景

API は結果が多い場合 `divideNumber` で分割返却する。現在の実装は最初のページのみ返却。

#### 変更点

- `cmd/search.go`, `cmd/diff.go` — `--page` フラグ（特定ページ指定）と `--all` フラグ（全ページ自動巡回）を追加
- `internal/api/houjin.go` — `SearchByName`, `GetDiff` に `divide` クエリパラメータを追加。`--all` 時はループで全ページ取得し Corporation 配列を結合
- `internal/model/corporation.go` — `Response` 構造体の `DivideNumber`, `DivideSize` フィールドを活用

#### 使用例

```bash
houjin search "トヨタ" --page 2        # 2ページ目のみ
houjin search "トヨタ" --all           # 全ページ巡回
houjin diff --from 2026-01-01 --to 2026-03-31 --all
```

---

### 1b) 法人種別フィルタ (`type` パラメータ)

#### 背景

API の `type` パラメータはビットマスクで法人種類をフィルタリングする。

#### 種別コード

| コード | 種別 |
|--------|------|
| `01` | 国の機関 |
| `02` | 地方公共団体 |
| `03` | 設立登記法人 |
| `04` | その他 |

#### 変更点

- `cmd/search.go` — `--type` フラグ追加。複数指定可能（例: `--type 03,04`）
- `internal/api/houjin.go` — `SearchByName` に `type` クエリパラメータを追加

#### 使用例

```bash
houjin search "トヨタ" --type 03          # 設立登記法人のみ
houjin search "トヨタ" --type 01,02       # 国の機関 + 地方公共団体
```

---

### 1c) 変更事由フィルタ (`kind` パラメータ in diff)

#### 背景

diff API の `kind` パラメータで変更種類をフィルタリングする。

#### 事由コード

| コード | 事由 |
|--------|------|
| `01` | 新規 |
| `02` | 商号又は名称の変更 |
| `03` | 国内所在地の変更 |
| `04` | 国外所在地の変更 |
| `11` | 登記記録の閉鎖等 |
| `12` | 登記記録の復活等 |
| `13` | 吸収合併 |
| `14` | 吸収合併無効 |
| `15` | 商号の登記の抹消 |
| `99` | 削除 |

#### 変更点

- `cmd/diff.go` — `--kind` フラグ追加。複数指定可能（例: `--kind 01,02`）
- `internal/api/houjin.go` — `GetDiff` に `kind` クエリパラメータを追加

#### 使用例

```bash
houjin diff --from 2026-01-01 --to 2026-03-31 --kind 01        # 新規のみ
houjin diff --from 2026-01-01 --to 2026-03-31 --kind 02,03     # 商号/所在地変更
```

---

### 1d) address パラメータ拡張

#### 背景

現在 `search` に `--pref`（都道府県コード）はあるが、`--city` の API 連携を確認・修正する。

#### 変更点

- `cmd/search.go` — `--city` フラグが API の `city` パラメータに正確にマッピングされるか確認・修正
- `internal/api/houjin.go` — `SearchByName` に `city` クエリパラメータを確実に渡す
- API 仕様: `city` は JIS X 0402 市区町村コード（3桁）。`pref` と併用

#### 使用例

```bash
houjin search "株式会社" --pref 13             # 東京都
houjin search "株式会社" --pref 13 --city 101  # 東京都千代田区
```

---

### 1e) バッチ入力

#### 背景

ファイルから法人番号リストを読み取り一括照会する機能。

#### 変更点

- `cmd/get.go` — `--file` フラグ追加
- テキストファイルから1行1法人番号で読み取り
- `-` 指定時は stdin から読み取り（パイプ対応）
- 各番号に対してチェックディジット検証後 API 呼び出し
- 結果は全法人を1つの配列に結合して出力

#### ファイル形式

```
# 法人番号リスト
2021001052596
2180301018771
```

1行に1法人番号。空行と `#` 始まりの行は無視。

#### 使用例

```bash
houjin get --file numbers.txt                    # ファイルから
cat numbers.txt | houjin get --file -            # stdinから
echo "2021001052596" | houjin get --file -       # パイプ
```

---

## Phase 2: App ID 申請ガイド

### 目的

README.md に App ID の申請方法と設定方法を記載する。

### 内容

- 国税庁法人番号公表サイトの Web-API 申請 URL
- 申請手順（無料、メールアドレス登録 → アプリケーション ID 発行）
- 発行後の設定方法3種:
  - 環境変数 `HOUJIN_APP_ID`
  - 設定ファイル `~/.config/houjin/config.yaml`
  - `--config` フラグ
- 設定確認方法の例示

### コード変更なし

ドキュメントのみ追加。

---

## 実装順序

| 順序 | Phase | ブランチ | リリース |
|------|-------|----------|----------|
| 1 | Phase 0 | `release/v0.1.0-prep` | v0.1.0 |
| 2 | Phase 1a | `feature/paging` | v0.2.0 に含める |
| 3 | Phase 1b | `feature/type-filter` | v0.2.0 に含める |
| 4 | Phase 1c | `feature/kind-filter` | v0.2.0 に含める |
| 5 | Phase 1d | `feature/address-params` | v0.2.0 に含める |
| 6 | Phase 1e | `feature/batch-input` | v0.2.0 に含める |
| 7 | Phase 2 | `docs/app-id-guide` | v0.2.0 に含める |

各 Phase は feature branch → PR ワークフローで進行。
