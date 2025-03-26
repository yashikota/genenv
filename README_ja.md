# genenv

`genenv` は、テンプレートファイルから.envファイルを生成するためのシンプルなツールです。

## 特徴

- `.env.example` のようなテンプレートファイルから `.env` ファイルを生成
- `${hoge}`のようなプレースホルダーを暗号学的に安全なランダム値に置き換え
- プレースホルダーではない箇所の値を保持
- 既存の`.env`ファイルを強制的に上書きするかどうかを選択可能
- カスタマイズ可能な出力ファイルパス
- 生成される値の長さと文字セットをカスタマイズ可能
- エスケープされたプレースホルダー（`\${not_a_placeholder}`）を適切に処理

## インストール

### ソースからのインストール

```bash
go install github.com/yashikota/genenv/cmd/genenv@latest
```

### バイナリからのインストール

[GitHub Releases](https://github.com/yashikota/genenv/releases)から最新のバイナリをダウンロードしてください。

## 使い方

```bash
# .env.exampleから.envファイルを生成
genenv .env.example

# カスタム出力パスで.envファイルを生成
genenv -o .env.production .env.example

# 既存の.envファイルを強制的に上書き
genenv -f .env.example

# カスタムの長さと文字セットで.envファイルを生成
genenv -l 32 -c numeric .env.example
```

### オプション

- `-f, --force`: 既存の`.env`ファイルを強制的に上書き
- `-o, --output`: 出力ファイルパスを指定（デフォルト: `.env`）
- `-i, --input`: コマンドライン引数の代わりにファイルからテンプレートを読み込む
- `-l, --length`: 生成されるランダム値の長さ（デフォルト: 24）
- `-c, --charset`: 生成される値の文字セット（デフォルト: alphanumeric）
  - 有効なオプション: `alphanumeric`, `alphabetic`, `uppercase`, `lowercase`, `numeric`
- `-h, --help`: ヘルプ情報を表示
- `-v, --version`: バージョン情報を表示

### 使用例

```bash
# 基本的な使い方
genenv .env.example

# カスタム出力パスで.envファイルを生成
genenv -o .env.production .env.example

# 既存の.envファイルを強制的に上書き
genenv -f .env.example

# ファイルからテンプレートを読み込む
genenv -i .env.example

# カスタムの長さ（32文字）で.envファイルを生成
genenv -l 32 .env.example

# カスタムの文字セット（数字のみ）で.envファイルを生成
genenv -c numeric .env.example

# カスタムの長さと文字セット（16文字の大文字）で.envファイルを生成
genenv -l 16 -c uppercase .env.example

# 複数のオプションを組み合わせる
genenv -i .env.example -o .env.production -f -l 32 -c alphabetic
```

`genenv .env.example`を実行すると、以下のような`.env`ファイルが生成されます：

```txt
# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
DB_PASSWORD=dGhpcyBpcyBhIHNlY3VyZSByYW5kb20gdmFsdWU=

# API設定
API_KEY=q7w8e9r0t1y2u3i4o5p6a7s8d9f0g1h2
API_URL=https://api.example.com
```

注意点：

- プレースホルダーのない値（`localhost`, `5432`, `https://api.example.com`）は保持されます
- プレースホルダー（`${db_user}`, `${db_password}`, `${api_key}`）は、一意の暗号学的に安全なランダム値に置き換えられます

### 長さと文字セットのカスタマイズ

`-l`と`-c`オプションを使用して、生成される値の長さと文字セットをカスタマイズできます。

```bash
# カスタムの長さ（32文字）で.envファイルを生成
genenv -l 32 .env.example

# カスタムの文字セット（数字のみ）で.envファイルを生成
genenv -c numeric .env.example

# カスタムの長さと文字セット（16文字の大文字）で.envファイルを生成
genenv -l 16 -c uppercase .env.example
```

### 文字セット

genenvは、生成される値に対して以下の文字セットをサポートしています：

- `alphanumeric`（デフォルト）: A-Z, a-z, 0-9
- `alphabetic`: A-Z, a-z
- `uppercase`: A-Z
- `lowercase`: a-z
- `numeric`: 0-9

### テンプレート形式

テンプレートファイルは、標準的な`.env`ファイルの形式で、生成すべき値にはプレースホルダーを使用します。

```txt
# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# API設定
API_KEY=${api_key}
API_URL=https://api.example.com

# その他の設定
DEBUG=true
SECRET_TOKEN=${secret_token}
CACHE_TTL=3600
```

この例では、`${db_user}`, `${db_password}`, `${api_key}`, `${secret_token}`が生成された値に置き換えられ、他の値は保持されます。

プレースホルダーを置き換えずにリテラルの`${...}`をテンプレートに含めるには、バックスラッシュでエスケープします：`\${not_a_placeholder}`。
