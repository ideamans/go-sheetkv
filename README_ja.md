# go-sheetkv

スプレッドシートをバックエンドとしたキーバリューストア (KVS) を提供する Go ライブラリです。Google Sheets と Excel ファイルの両方に対応しています。

## 特徴

- Google Sheets と Excel を KVS として利用
- メモリキャッシュによる高速アクセス
- 自動同期機能
- 型安全な API
- リトライ機能付き
- 複数の認証方式をサポート（Google Sheets）
- 認証不要でローカルファイル操作（Excel）

## 重要な注意事項

⚠️ **このパッケージはシンプルなバッチ処理での利用を想定しており、複数のプロセスからの同時アクセスには対応していません。** すべてのデータは各プロセス内のメモリにキャッシュされ、プロセス間の同期機構はありません。複数のプロセスから同時にこのパッケージを使用すると、データの不整合が発生する可能性があります。

## インストール

```bash
go get github.com/ideamans/go-sheetkv
```

## 使い方

```go
import (
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
    "github.com/ideamans/go-sheetkv/adapters/excel"
)
```

### 基本的な使用例

```go
package main

import (
    "context"
    "log"
    "time"
    
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/googlesheets"
)

func main() {
    ctx := context.Background()
    
    // アダプターの設定と作成
    adapterConfig := googlesheets.Config{
        SpreadsheetID: "your-spreadsheet-id",
        SheetName:     "users",
    }
    adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./credentials.json")
    if err != nil {
        log.Fatal(err)
    }

    // クライアントの設定と作成
    // Google Sheets用の推奨デフォルト設定を使用
    clientConfig := googlesheets.DefaultClientConfig()
    // 必要に応じてカスタマイズ
    // clientConfig.SyncInterval = 30 * time.Second
    
    client := sheetkv.New(adapter, clientConfig)
    
    // 初期データの読み込み
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // レコードの追加
    record := &sheetkv.Record{
        Values: map[string]interface{}{
            "name": "山田太郎",
            "age":  25,
        },
    }
    err = client.Append(record)
    if err != nil {
        log.Fatal(err)
    }

    // レコードの検索
    results, err := client.Query(sheetkv.Query{
        Conditions: []sheetkv.Condition{
            {Column: "age", Operator: ">=", Value: 20},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // 結果の表示
    for _, r := range results {
        name := r.GetAsString("name", "")
        age := r.GetAsInt64("age", 0)
        log.Printf("Row %d: %s (age: %d)", r.Key, name, age)
    }
}
```

### Excel を使用する例

```go
package main

import (
    "context"
    "log"
    "time"
    
    sheetkv "github.com/ideamans/go-sheetkv"
    "github.com/ideamans/go-sheetkv/adapters/excel"
)

func main() {
    // Excel アダプターの設定と作成（認証不要）
    adapterConfig := &excel.Config{
        FilePath:  "./data.xlsx",
        SheetName: "users",
    }
    adapter, err := excel.New(adapterConfig)
    if err != nil {
        log.Fatal(err)
    }

    // クライアントの作成と初期化
    // Excel用の推奨デフォルト設定を使用
    client := sheetkv.New(adapter, excel.DefaultClientConfig())
    
    ctx := context.Background()
    if err := client.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 以降の操作は Google Sheets と同じ
}
```

## 認証方式

### Google Sheets の認証

### 1. サービスアカウント JSON ファイル

```go
adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "./service-account.json")
```

### 2. 環境変数経由

```go
// GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
adapter, err := googlesheets.NewWithJSONKeyFile(ctx, adapterConfig, "")
```

### 3. サービスアカウントキー直接指定

```go
adapter, err := googlesheets.NewWithServiceAccountKey(
    ctx, 
    adapterConfig,
    "service-account@project.iam.gserviceaccount.com",
    "-----BEGIN PRIVATE KEY-----\n...",
)
```

## データ型

Record の Values は `map[string]interface{}` 型ですが、型安全なアクセスのためのヘルパーメソッドが提供されています：

```go
// Getter メソッド
name := record.GetAsString("name", "デフォルト値")
age := record.GetAsInt64("age", 0)
price := record.GetAsFloat64("price", 0.0)
active := record.GetAsBool("active", false)
tags := record.GetAsStrings("tags", []string{})
created := record.GetAsTime("created_at", time.Now())

// Setter メソッド
record.SetString("name", "新しい名前")
record.SetInt64("age", 30)
record.SetFloat64("price", 1980.0)
record.SetBool("active", true)
record.SetStrings("tags", []string{"tag1", "tag2"})
record.SetTime("updated_at", time.Now())
```

## クエリ

複数の条件を組み合わせた検索が可能です：

```go
results, err := client.Query(sheetkv.Query{
    Conditions: []sheetkv.Condition{
        {Column: "status", Operator: "==", Value: "active"},
        {Column: "age", Operator: ">=", Value: 18},
        {Column: "age", Operator: "<=", Value: 65},
        {Column: "role", Operator: "in", Value: []interface{}{"admin", "user"}},
    },
    Limit:  10,
    Offset: 0,
})
```

### サポートされる演算子

- `==` : 等しい
- `!=` : 等しくない
- `>` : より大きい
- `>=` : 以上
- `<` : より小さい
- `<=` : 以下
- `in` : 含まれる（配列で値を指定）
- `between` : 範囲内（2要素の配列で範囲を指定）

## スプレッドシートの構造

- 1行目: カラム名（スキーマ定義）
- 2行目以降: データ
- キーは行番号（2から開始）

## 同期戦略

本ライブラリは2種類の同期戦略を実装しています：

### 欠番維持同期（定期同期時のデフォルト）
- 削除されたレコードは空行として同期されます
- メモリ上の行番号（キー）とスプレッドシート上の行番号の一致を維持します
- 新規レコード追加時も、既存の最大キーから連続的に番号が振られます
- 定期的な同期処理で自動的に使用されます

### コンパクト化同期（Close時に使用）
- 削除されたレコードは取り除かれ、データが詰めて配置されます
- 空行を削除することでスプレッドシートのサイズを最適化します
- 同期後はスプレッドシート上の行番号とレコードのキーが一致しない場合があります
- 末尾の余分な行も自動的に削除され、クリーンなデータを維持します
- `Close()` メソッド呼び出し時に自動的に使用されます

## 開発

### テストの実行

```bash
# 単体テスト
make test-unit

# 統合テスト（要 .env 設定）
make test-integration

# API テスト（要 .env 設定）
make test-api

# 全テスト
make test
```

### 必要な環境変数

テスト実行には `.env` ファイルが必要です：

```env
# Google Sheets を使用する場合
GOOGLE_APPLICATION_CREDENTIALS=./service-account.json
TEST_GOOGLE_SHEET_ID=your-test-spreadsheet-id

# 追加の認証方法（オプション）
TEST_CLIENT_EMAIL=service-account@project.iam.gserviceaccount.com
TEST_CLIENT_PRIVATE_KEY=-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----

# 注意:
# - シート名は自動的に設定されます（integration/api）
# - Google Sheets の認証情報が設定されていない場合、自動的に Excel アダプターでテストが実行されます
# - Excel アダプターは常にテストされます
```

### CI/CD

GitHub Actions で Google Sheets を使用したテストを実行する場合は、リポジトリに以下の Secrets を設定してください：

- `SERVICE_ACCOUNT_JSON`: サービスアカウントの JSON ファイル内容
- `TEST_CLIENT_EMAIL`: サービスアカウントのメールアドレス
- `TEST_CLIENT_PRIVATE_KEY`: サービスアカウントの秘密鍵
- `TEST_GOOGLE_SHEET_ID`: テスト用の Google スプレッドシート ID

詳細は [.github/CI_SECRETS.md](.github/CI_SECRETS.md) を参照してください。

## ライセンス

MIT License