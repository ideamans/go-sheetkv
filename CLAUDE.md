# sheetkv 設計書

## 0. はじめに

- Claude Code でコミットや push は試みないこと

## 1. 概要

### 1.1 目的

Google Spreadsheet を Key-Value Store (KVS)として利用するための Go 言語ライブラリを提供する。

### 1.2 主な特徴

- Google Spreadsheet をバックエンドとした KVS 操作
- メモリキャッシュによる API Quota 回避
- 複数のスプレッドシートバックエンドへの対応を考慮した設計（将来的に Excel, CSV 等も追加可能）
- 柔軟な認証方式のサポート
- MIT ライセンスでの OSS 提供

### 1.3 重要な制限事項

⚠️ **このパッケージはシンプルなバッチ処理での利用を想定しており、複数のプロセスからの同時アクセスには対応していません。**

- すべてのデータは各プロセス内のメモリにキャッシュされる
- プロセス間の同期機構は存在しない
- 複数のプロセスから同時にアクセスした場合、データの不整合が発生する可能性がある
- 単一プロセス内での並行処理は、内部のmutexにより安全に制御される

## 2. アーキテクチャ

### 2.1 レイヤー構成

```
┌────────────────────────────────────────────────────┐
│              Application Layer                      │
├────────────────────────────────────────────────────┤
│              Client Interface Layer                │
├────────────────────────────────────────────────────┤
│              Cache Layer (Memory)                  │
├────────────────────────────────────────────────────┤
│               Adapter Interface Layer              │
├────────────────────────────────────────────────────┤
│   Google Sheets │ Excel │ CSV │ ...               │
└────────────────────────────────────────────────────┘
```

### 2.2 主要コンポーネント

#### 2.2.1 Application Layer

- ユーザー向けの Go 言語 API
- 直接的なライブラリ利用

#### 2.2.2 Client Interface Layer

- 共通の KVS 操作インターフェース
- CRUD 操作の提供
- キャッシュとアダプターの統合管理

#### 2.2.3 Cache Manager

- インメモリでのデータ管理
- 定期的な同期処理
- 変更追跡

#### 2.2.4 Adapter Interface Layer

- スプレッドシートバックエンドの抽象化
- 各バックエンド固有の実装を統一インターフェースで提供
- Google Sheets、Excel、CSV 等の実装を可能にする

#### 2.2.5 Sync Manager

- バックエンドへの同期処理
- リトライ機構
- API Quota 管理

## 3. データモデル

### 3.1 スプレッドシート構造

- **シート**: テーブルに相当
- **1 行目**: スキーマ定義（カラム名）
- **2 行目以降**: データレコード
- **行番号**: キー（Primary Key）として使用（2 から開始）

### 3.2 レコード構造

```go
type Record struct {
    Key    int                    // 行番号（2から開始、1行目はカラム定義）
    Values map[string]interface{} // カラム名と値のマップ
}

// 型変換ヘルパーメソッド（Getter）
func (r *Record) GetAsString(col string, defaultValue string) string
func (r *Record) GetAsInt64(col string, defaultValue int64) int64
func (r *Record) GetAsFloat64(col string, defaultValue float64) float64
func (r *Record) GetAsStrings(col string, defaultValue []string) []string
func (r *Record) GetAsBool(col string, defaultValue bool) bool
func (r *Record) GetAsTime(col string, defaultValue time.Time) time.Time

// 型変換ヘルパーメソッド（Setter）
func (r *Record) SetString(col string, value string)
func (r *Record) SetInt64(col string, value int64)
func (r *Record) SetFloat64(col string, value float64)
func (r *Record) SetStrings(col string, value []string)
func (r *Record) SetBool(col string, value bool)
func (r *Record) SetTime(col string, value time.Time)
```

## 4. API 設計

### 4.1 パッケージ構造

```go
import (
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
    "github.com/ideamans/go-sheetkv/adapters/excel"
    // 将来的な拡張
    // "github.com/ideamans/go-sheetkv/adapters/csv"
)
```

### 4.2 アダプター初期化

#### Google Sheets アダプター

```go
// アダプター設定
type googlesheets.Config struct {
    SpreadsheetID string
    SheetName     string
}

// JSON認証での初期化
adapter, err := googlesheets.NewWithJSONKeyFile(ctx, config, "./service-account.json")
// jsonPath が空の場合は GOOGLE_APPLICATION_CREDENTIALS 環境変数を参照

// JSONデータでの初期化
adapter, err := googlesheets.NewWithJSONKeyData(ctx, config, []byte(jsonData))

// キー認証での初期化
adapter, err := googlesheets.NewWithServiceAccountKey(ctx, config, email, privateKey)
```

#### Excel アダプター

```go
// アダプター設定
type excel.Config struct {
    FilePath  string // Excelファイルのパス
    SheetName string // シート名
}

// Excel アダプターの初期化（認証不要）
adapter, err := excel.New(config)
```

### 4.3 クライアント初期化

```go
// クライアント設定
type Config struct {
    SyncInterval  time.Duration  // デフォルト: 30秒
    MaxRetries    int           // デフォルト: 3
    RetryInterval time.Duration  // デフォルト: 1秒（指数バックオフ）
}

// クライアントの作成
client := sheetkv.New(adapter, config)
// config が nil の場合はデフォルト値を使用

// 初期化（既存データの読み込み）
err := client.Initialize(ctx)

// 終了処理（同期を保証）
err := client.Close()
```

### 4.4 CRUD 操作

```go
// レコード追加（キーは自動採番）
func (c *Client) Append(record *Record) error

// レコード検索
type Condition struct {
    Column   string      // カラム名
    Operator string      // 演算子: ==, !=, >, >=, <, <=, in, between
    Value    interface{} // 比較値（inの場合は[]interface{}, betweenの場合は[2]interface{}）
}

type Query struct {
    Conditions []Condition // AND条件として評価
    Limit      int
    Offset     int
}

func (c *Client) Query(query Query) ([]*Record, error)

// キーによる取得（行番号を指定）
func (c *Client) Get(key int) (*Record, error)

// キーによる上書き（行番号を指定）
func (c *Client) Set(key int, record *Record) error

// キーによる部分更新（行番号を指定）
func (c *Client) Update(key int, updates map[string]interface{}) error

// キーによる削除（行番号を指定）
func (c *Client) Delete(key int) error

// 強制同期
func (c *Client) Sync() error
```

## 5. 内部設計

### 5.1 Cache Layer

```go
type Cache struct {
    mu       sync.RWMutex
    data     map[int]*Record // Key -> Record (row number)
    dirty    map[int]bool    // 変更追跡
    schema   []string
}
```

### 5.2 Adapter Interface

```go
type Adapter interface {
    Load(ctx context.Context) ([]*Record, []string, error)
    Save(ctx context.Context, records []*Record, schema []string) error
    BatchUpdate(ctx context.Context, operations []Operation) error
}

type Operation struct {
    Type   OperationType // Add, Update, Delete
    Record *Record
}

type OperationType int

const (
    OperationAdd OperationType = iota
    OperationUpdate
    OperationDelete
)
```

### 5.3 Sync Manager

```go
type SyncManager struct {
    cache    *Cache
    adapter  Adapter
    interval time.Duration
    ticker   *time.Ticker
    done     chan bool
}
```

## 6. 同期戦略

### 6.1 全量同期方式

- メモリ上の全データを Spreadsheet に反映
- カラムと行の増減に対応
- データの整合性を保証

### 6.2 同期アルゴリズム

#### カラム同期

1. 現在のメモリ上のスキーマ（カラム順序）を取得
2. Spreadsheet の 1 行目に現在のスキーマを書き込み
3. 既存カラムの順序を維持しつつ、新規カラムは末尾に追加
4. メモリ上のカラム数より後ろのカラムは削除

#### データ同期

1. メモリ上の全レコードをキーでソート
2. 2 行目から順番に全レコードを書き込み
3. メモリ上のレコード数より後ろの行は削除
4. 空白セルも明示的にクリア

### 6.3 同期処理の実装

```go
func (c *Client) saveToAdapter(ctx context.Context) error {
    // dirtyなレコードがある場合のみ同期
    dirtyKeys := c.cache.GetDirtyKeys()
    if len(dirtyKeys) == 0 {
        return nil
    }

    // 全レコードとスキーマを取得
    records := c.cache.GetAllRecords()
    schema := c.cache.GetSchema()

    // アダプターに保存（リトライ付き）
    var err error
    for i := 0; i <= c.config.MaxRetries; i++ {
        err = c.adaptor.Save(ctx, records, schema)
        if err == nil {
            c.cache.ClearDirty()
            return nil
        }

        if i < c.config.MaxRetries {
            backoff := time.Duration(1<<uint(i)) * c.config.RetryInterval
            time.Sleep(backoff)
        }
    }

    return fmt.Errorf("failed to save after %d retries: %w", c.config.MaxRetries, err)
}

// カラムのマージ（既存順序を維持）
func mergeSchemas(current, sheet []string) []string {
    result := make([]string, 0)
    seen := make(map[string]bool)

    // 既存のシートカラムの順序を維持
    for _, col := range sheet {
        if contains(current, col) {
            result = append(result, col)
            seen[col] = true
        }
    }

    // 新規カラムを末尾に追加
    for _, col := range current {
        if !seen[col] {
            result = append(result, col)
        }
    }

    return result
}
```

### 6.4 削除処理の詳細

```go
// 不要な行・列の削除リクエスト生成
func createCleanupRequests(dataRows int, dataCols int) []BatchUpdateRequest {
    requests := []BatchUpdateRequest{}

    // 行の削除
    // データ行数 + 1（スキーマ行）より後の行を削除
    deleteRowsRequest := DeleteRangeRequest{
        Range: CellRange{
            StartRowIndex: dataRows + 1,
            EndRowIndex:   MAX_ROWS, // シートの最大行数
        },
    }
    requests = append(requests, deleteRowsRequest)

    // 列の削除
    // 使用カラム数より後の列を削除
    deleteColsRequest := DeleteRangeRequest{
        Range: CellRange{
            StartColumnIndex: dataCols,
            EndColumnIndex:   MAX_COLS, // シートの最大列数
        },
    }
    requests = append(requests, deleteColsRequest)

    return requests
}
```

### 6.5 同期の最適化

#### バッチ処理による効率化

- 複数の更新操作を 1 回の API リクエストにまとめる
- セル範囲の更新を最小限に抑える
- 不要な空白セルのクリアも含める

#### エラー処理とリトライ

```go
func (s *SyncManager) syncWithRetry() error {
    maxRetries := s.config.MaxRetries
    backoff := time.Second

    for i := 0; i < maxRetries; i++ {
        err := s.syncToSheet()
        if err == nil {
            return nil
        }

        // API Quota エラーの場合は待機時間を延長
        if isQuotaError(err) {
            backoff *= 2
        }

        time.Sleep(backoff)
    }

    return fmt.Errorf("sync failed after %d retries", maxRetries)
}
```

### 6.6 同期制御

#### 排他制御

定期同期では、前回の同期がまだ実行中の場合、その回の同期をスキップします：

```go
func (sm *SyncManager) performSync() {
    // 排他制御のためのロック取得を試行
    if !sm.syncMutex.TryLock() {
        // 前回の同期がまだ実行中なのでスキップ
        return
    }
    defer sm.syncMutex.Unlock()

    // 同期処理の実行
    _ = sm.client.saveToAdaptor()
}
```

#### Close 時の強制同期

`Close()` メソッドでは、進行中の同期が完了するまで待機し、最新データの同期を保証します：

```go
func (c *Client) Close() error {
    // 同期マネージャーを停止
    if c.syncManager != nil {
        c.syncManager.Stop() // 進行中の同期が完了するまで待機
    }

    // 最終同期を実行
    if err := c.saveToAdaptor(); err != nil {
        return fmt.Errorf("failed to sync on close: %w", err)
    }

    return nil
}
```

### 7.1 エラー種別

```go
var (
    ErrKeyNotFound    = errors.New("key not found")
    ErrDuplicateKey   = errors.New("duplicate key")
    ErrSyncFailed     = errors.New("sync failed")
    ErrQuotaExceeded  = errors.New("quota exceeded")
)
```

### 7.2 リトライ戦略

#### 自動リトライ

全ての Google Sheets API 呼び出しは自動的にリトライされます：

- **指数バックオフ**: 1 秒、2 秒、4 秒...と待機時間が増加
- **最大リトライ回数**: `Config.MaxRetries`で設定（デフォルト: 3）
- **対象エラー**: ネットワークエラー、一時的な API 障害（503 エラーなど）

```go
// リトライ動作の例
for i := 0; i <= config.MaxRetries; i++ {
    err = apiCall()
    if err == nil {
        break
    }

    if i < config.MaxRetries {
        backoff := time.Duration(1<<uint(i)) * time.Second
        time.Sleep(backoff)
    }
}
```

#### エラーの種類

- **一時的エラー**: 自動的にリトライ
- **永続的エラー**: 即座に失敗（認証エラー、権限不足など）
- **Quota エラー**: より長い待機時間でリトライ

## 8. 使用例

### 8.1 基本的な使用方法

```go
import (
    "context"
    "time"

    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
)

ctx := context.Background()

// アダプターの設定と初期化
adapterConfig := googlesheets.Config{
    SpreadsheetID: "your-spreadsheet-id",
    SheetName:     "users",
}

// 方法1: JSONファイル認証
adapter := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./service-account.json")

// 方法2: 環境変数GOOGLE_APPLICATION_CREDENTIALSを使用
adapter := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "")

// 方法3: サービスアカウントキー認証
adapter := googlesheets.NewWithServiceAccountKey(
    ctx,
    adapterConfig,
    "service-account@project.iam.gserviceaccount.com",
    "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
)

// クライアントの設定と作成
// 推奨デフォルト設定を使用
clientConfig := googlesheets.DefaultClientConfig()
// デフォルト値: SyncInterval=10秒, MaxRetries=3, RetryInterval=20秒

// またはカスタム設定
clientConfig := &sheetkv.Config{
    SyncInterval:  30 * time.Second,
    MaxRetries:    3,
    RetryInterval: 1 * time.Second,
}

client := sheetkv.New(adapter, clientConfig)
defer client.Close() // 必ず同期を完了させる
```

### 8.2 Excel アダプターの使用例

```go
import (
    "context"

    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/excel"
)

// Excel アダプターの設定（認証不要）
adapterConfig := &excel.Config{
    FilePath:  "./data.xlsx",  // Excelファイルのパス
    SheetName: "users",        // シート名
}

adapter, err := excel.New(adapterConfig)
if err != nil {
    log.Fatal(err)
}

// クライアントの作成
// Excel用の推奨デフォルト設定を使用
client := sheetkv.New(adapter, excel.DefaultClientConfig())
// デフォルト値: SyncInterval=1秒, MaxRetries=3, RetryInterval=5秒

// 初期化
ctx := context.Background()
err = client.Initialize(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 以降の操作はGoogle Sheetsと完全に同じ
```

### 8.3 CRUD 操作の例

```go
// レコード追加（新しい行として追加され、自動的に行番号が割り当てられる）
user := sheetkv.Record{
    Values: map[string]interface{}{
        "name":  "山田太郎",
        "email": "yamada@example.com",
        "age":   30,
    },
}
err = client.Append(user)

// レコード検索（複数条件AND）
results, err := client.Query(sheetkv.Query{
    Conditions: []sheetkv.Condition{
        {Column: "age", Operator: ">=", Value: 20},
        {Column: "age", Operator: "<=", Value: 40},
        {Column: "status", Operator: "==", Value: "active"},
        {Column: "role", Operator: "in", Value: []interface{}{"admin", "moderator"}},
    },
    Limit: 10,
})

// レコード更新（行番号2のレコードを更新）
err = client.Update(2, map[string]interface{}{
    "email": "new-email@example.com",
    "updated_at": time.Now(),
})

// 型安全な操作の例
record, err := client.Get(2)
if err == nil {
    // Getterの使用例
    name := record.GetAsString("name", "名無し")
    age := record.GetAsInt64("age", 0)
    tags := record.GetAsStrings("tags", []string{})

    // Setterの使用例
    record.SetString("status", "active")
    record.SetInt64("login_count", 10)
    record.SetStrings("tags", []string{"premium", "verified"})

    err = client.Set(2, record)
}
```

## 9. パッケージ構造

```
github.com/ideamans/go-sheetkv/
├── client.go            # Client実装
├── record.go            # Record構造体と型変換メソッド
├── query.go             # Query構造体とクエリ評価
├── cache.go             # キャッシュ実装
├── sync.go              # 同期マネージャー
├── adapter.go           # Adapterインターフェース定義
├── errors.go            # エラー定義
├── config.go            # Config構造体
├── adapters/
│   ├── googlesheets/    # Google Sheets実装
│   │   ├── sheets.go    # Adapterインターフェース実装
│   │   ├── auth.go      # 認証関連
│   │   └── config.go    # Google Sheets固有の設定
│   ├── excel/           # Excel実装（実装済み）
│   │   ├── excel.go     # Adapterインターフェース実装
│   │   ├── config.go    # Excel固有の設定
│   │   └── errors.go    # エラー定義
│   └── csv/             # CSV実装（将来）
├── example/             # 使用例
│   ├── main.go         # Google Sheets の例
│   └── excel/          # Excel の例
│       ├── main.go
│       └── example_data.xlsx
├── examples/            # テスト用例
│   └── integration-test/
│       └── example_test.go
├── tests/
│   ├── common/         # 共通テストユーティリティ
│   │   └── adapter_test_suite.go
│   ├── integration/    # 結合テスト
│   │   ├── adapter_integration_test.go
│   │   └── README.md
│   └── api/            # APIテスト
│       ├── adapter_api_test.go
│       └── README.md
├── LICENSE             # MITライセンス
├── README.md           # 英語版README
├── README_ja.md        # 日本語版README
├── CLAUDE.md          # 設計書
└── go.mod
```

## 10. 実装状況

### 10.1 実装済み機能

- ✅ Google Sheets アダプター（完全実装）
  - JSON ファイル認証
  - JSON データ認証
  - サービスアカウントキー認証
  - リトライ機能付き API 呼び出し
- ✅ Excel アダプター（完全実装）
  - xlsx ファイル形式対応
  - ローカルファイル読み書き
  - 認証不要
- ✅ メモリキャッシュ層
  - 高速アクセス
  - 変更追跡（dirty tracking）
  - スキーマ管理
- ✅ 同期マネージャー
  - 定期同期
  - 排他制御
  - 強制同期
- ✅ クエリシステム
  - 複数条件の AND 検索
  - 各種演算子サポート
  - ページネーション
- ✅ 型安全 API
  - Getter/Setter メソッド
  - 型変換サポート
- ✅ テストスイート
  - 単体テスト
  - 統合テスト（アダプター共通）
  - API テスト（アダプター共通）

## 11. 今後の拡張性

### 11.1 対応バックエンド

#### 実装済み

- **Google Sheets** (完全対応)

  - 認証: サービスアカウント、OAuth2
  - リアルタイム同期
  - API 経由でのアクセス
  - 推奨デフォルト設定:
    - SyncInterval: 10 秒
    - MaxRetries: 3 回
    - RetryInterval: 20 秒

- **Microsoft Excel** (xlsx 形式、実装済み)
  - 使用ライブラリ: `github.com/xuri/excelize/v2`
  - ローカルファイルの読み書き
  - 認証不要
  - xlsx 形式のみサポート（xls 非対応）
  - パスワード保護非対応
  - 推奨デフォルト設定:
    - SyncInterval: 1 秒
    - MaxRetries: 3 回
    - RetryInterval: 5 秒

#### 対応予定

- CSV ファイル
- その他のスプレッドシート形式

### 11.2 機能拡張案

- トランザクション対応
- インデックス機能
- スキーマバリデーション
- データ型の自動変換
- 複数シート（テーブル）の JOIN 操作

## 12. 性能考慮事項

### 12.1 メモリ使用量

- レコード数に比例したメモリ使用
- 大規模データセット用のページング機能

### 12.2 同期性能

- バッチ処理による効率化
- 差分同期による通信量削減

## 13. 型変換の詳細設計

### 13.1 型変換ルール

#### String 型への変換

- 文字列: そのまま返す
- 数値: fmt.Sprintf で文字列化
- bool: "true" または "false"
- []string: カンマ区切りで結合
- その他: fmt.Sprintf("%v") で文字列化

#### Int64 型への変換

- 文字列: strconv.ParseInt でパース
- float64: int64 にキャスト
- その他: デフォルト値を返す

#### Float64 型への変換

- 文字列: strconv.ParseFloat でパース
- int 系: float64 にキャスト
- その他: デフォルト値を返す

#### []string 型への変換

- 文字列: カンマで分割
- []interface{}: 各要素を文字列化
- その他: デフォルト値を返す

### 13.2 実装例

```go
// GetAsString の実装例
func (r *Record) GetAsString(col string, defaultValue string) string {
    v, ok := r.Values[col]
    if !ok {
        return defaultValue
    }

    switch val := v.(type) {
    case string:
        return val
    case int, int64, float64:
        return fmt.Sprintf("%v", val)
    case bool:
        if val {
            return "true"
        }
        return "false"
    case []string:
        return strings.Join(val, ",")
    default:
        return fmt.Sprintf("%v", val)
    }
}

// SetStrings の実装例（スプレッドシートでの配列表現）
func (r *Record) SetStrings(col string, value []string) {
    // カンマ区切りで保存
    r.Values[col] = strings.Join(value, ",")
}
```

### 13.3 スプレッドシートでの型表現

Google Spreadsheet では全ての値が文字列として保存されるため、以下の変換規則を適用：

- **数値**: そのまま数値として保存
- **文字列**: そのまま文字列として保存
- **配列**: カンマ区切り文字列として保存（例: "tag1,tag2,tag3"）
- **日時**: ISO 8601 形式の文字列として保存

## 14. クエリシステムの詳細設計

### 14.1 クエリ演算子

| 演算子  | 説明       | 値の型         | 使用例                                                                    |
| ------- | ---------- | -------------- | ------------------------------------------------------------------------- |
| ==      | 等しい     | any            | `{Column: "status", Operator: "==", Value: "active"}`                     |
| !=      | 等しくない | any            | `{Column: "status", Operator: "!=", Value: "deleted"}`                    |
| >       | より大きい | number         | `{Column: "age", Operator: ">", Value: 18}`                               |
| >=      | 以上       | number         | `{Column: "score", Operator: ">=", Value: 60}`                            |
| <       | より小さい | number         | `{Column: "price", Operator: "<", Value: 1000}`                           |
| <=      | 以下       | number         | `{Column: "stock", Operator: "<=", Value: 10}`                            |
| in      | 含まれる   | []interface{}  | `{Column: "role", Operator: "in", Value: []interface{}{"admin", "user"}}` |
| between | 範囲内     | [2]interface{} | `{Column: "age", Operator: "between", Value: [2]interface{}{20, 30}}`     |

### 14.2 クエリ評価ルール

- 全ての条件は AND で結合される
- 条件は定義された順序で評価される
- 型の不一致は false として評価される
- 空の条件配列は全レコードにマッチする

### 14.3 クエリ使用例

```go
// 複雑なクエリの例
query := Query{
    Conditions: []Condition{
        {Column: "age", Operator: "between", Value: [2]interface{}{20, 40}},
        {Column: "status", Operator: "==", Value: "active"},
        {Column: "department", Operator: "in", Value: []interface{}{"sales", "marketing"}},
        {Column: "salary", Operator: ">=", Value: 50000},
    },
    Limit:  20,
    Offset: 0,
}
```
