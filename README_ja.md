# genenv

`genenv` は、テンプレートファイルから `.env` ファイルを生成し、プレースホルダーに暗号化された安全なランダム値を自動的に入力するシンプルなCLIツールです。

*Read this in [English](README.md)*

## 特徴

- `.env.example` のようなテンプレートファイルから `.env` ファイルを生成
- `${hoge}`のようなプレースホルダーを暗号学的に安全なランダム値に置き換え
- プレースホルダーではない箇所の値を保持
- 既存の`.env`ファイルを強制的に上書きするかどうかを選択可能
- カスタマイズ可能な出力ファイルパス
- 生成される値の長さと文字セットをカスタマイズ可能
- エスケープされたプレースホルダー（`\${not_a_placeholder}`）を適切に処理
- フィールド検証とタイプチェック機能を備えたインタラクティブモード
- 既存の`.env`ファイルと比較して新しいフィールドのみを追加する機能
- `.env`ファイルに既に存在するフィールドをスキップするオプション
- テンプレートなしで新しい`.env`ファイルを最初から作成する機能
- ローカルIPアドレス（IPv4およびIPv6）の自動検出と挿入

## サンプル例

`examples`ディレクトリには、`genenv`の機能を示す様々な例が含まれています：

- **Basic**: 基本的なプレースホルダーの置き換え
- **With Metadata**: フィールドのメタデータと検証
- **With Types**: フィールドタイプの検証
- **Auto IP**: ローカルIPアドレスの自動検出と挿入
- **Complex**: 複雑な場合の例
- **Escaped Placeholders**: リテラル`${...}`構文の保持
- **Compare with Existing**: 既存の`.env`ファイルへの新しいフィールドの追加
- **Custom Character Sets**: 異なる文字セットと長さの使用
- **New From Scratch**: テンプレートなしで新しい`.env`ファイルを作成

各例には以下が含まれます：

- `.env.example`テンプレートファイル
- 期待される出力を示すサンプル`.env`ファイル
- 機能と使用方法を説明する詳細なREADME

### サンプルの実行

`example`のディレクトリに移動して実行します：

```bash
cd example
genenv .env.example
```

### サンプルのテスト

提供されているテストスクリプトを使用して、すべての例の自動テストを実行できます：

```bash
# examplesディレクトリに移動
cd examples

# まずgenenvをビルド（まだビルドされていない場合）
go build -o genenv ../cmd/genenv/main.go

# テストを実行
go run test_examples.go
```

## インストール

### ソースからのインストール

```bash
go install github.com/yashikota/genenv/cmd/genenv@latest
```

### バイナリからのインストール

[GitHub Releases](https://github.com/yashikota/genenv/releases)から最新のバイナリをダウンロードしてください。

## 使い方

基本的な使い方：

```bash
# .env.exampleから.envファイルを生成
genenv .env.example
```

インタラクティブモード（引数なし）：

```bash
genenv
```

これにより、すべての設定オプションとフィールド値の入力を促すインタラクティブモードが開始されます。

### オプション

- `-f, --force`: 既存の`.env`ファイルを強制的に上書き
- `-o, --output`: 出力ファイルパスを指定（デフォルト: `.env`）
- `-i, --input`: コマンドライン引数の代わりにファイルからテンプレートを読み込む
- `-l, --length`: 生成されるランダム値の長さ（デフォルト: 24）
- `-c, --charset`: 生成される値の文字セット（デフォルト: alphanumeric）
  - 有効なオプション: `alphanumeric`, `alphabetic`, `uppercase`, `lowercase`, `numeric`
- `-I, --interactive`: インタラクティブモードで実行し、値の入力を促す
- `-C, --compare`: 既存の`.env`ファイルと比較して新しいフィールドのみを追加
- `-S, --skip-existing`: 既に`.env`ファイルに存在するフィールドをスキップ
- `-y, --yes`: すべてのプロンプトに対して自動的にyesと回答
- `-n, --no`: すべてのプロンプトに対して自動的にnoと回答
- `-h, --help`: ヘルプ情報を表示
- `-v, --version`: バージョン情報を表示

### 使用例

`.env.example`ファイルが以下のような場合：

```txt
# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# API設定
API_KEY=${api_key}
API_URL=https://api.example.com
```

`genenv .env.example`を実行すると、以下のような`.env`ファイルが生成されます：

```txt
# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
DB_PASSWORD=dGhpcyBpcyBhIHNlY3VyZSByYW5kb20gdmFsdWU

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

### インタラクティブモード

引数なしで`genenv`を実行するか、`-I/--interactive`フラグを使用すると、インタラクティブモードが開始されます。このモードでは：

1. すべての設定オプション（テンプレートファイル、出力ファイルなど）の入力を促されます
2. テンプレート内の各フィールドに対して値の入力を促されます
3. フィールドの入力プロンプトには以下の情報が含まれます：
   - フィールドの説明（利用可能な場合）
   - フィールドが必須か任意か
   - フィールドの型（string、int、boolなど）
   - デフォルト値（利用可能な場合）
   - 現在の値（現在の.envファイルに存在する場合）
   - IPフィールドの場合、検出されたIPアドレス（利用可能な場合）

インタラクティブモードの例：

```bash
$ genenv

Welcome to genenv interactive mode!
Press Enter to use default values.
Enter template file path (.env.example): 
Enter output file path (.env): 
Enter length for generated values (24): 
Enter charset for generated values (alphanumeric, alphabetic, uppercase, lowercase, numeric) [alphanumeric]: 
Compare with existing .env file? (y/N): y
Skip fields that already exist in the .env file? (y/N): y
Force overwrite of existing .env file? (y/N): y

Configuration summary:
Template file: .env.example
Output file: .env
Value length: 24
Charset: alphanumeric
Compare with existing .env: true
Skip existing fields: true
Force overwrite: true

DB_HOST (データベースホスト, [OPTIONAL], type: string): localhost
DB_PORT (データベースポート, [OPTIONAL], type: int): 5432
DB_USER (データベースユーザー名, [REQUIRED], type: string): admin
DB_PASSWORD (データベースパスワード, [REQUIRED], type: string): 
Generated random value: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
API_KEY (外部サービス用APIキー, [REQUIRED], type: string): 
Generated random value: q7w8e9r0t1y2u3i4o5p6a7s8d9f0g1h2
SERVER_IP (サーバーIPアドレス, [REQUIRED], type: ip) (検出: 192.168.1.100): 
検出されたIPを使用: 192.168.1.100

Successfully generated .env from .env.example
```

### テンプレートなしで新しい.envファイルを作成

`-N/--new`フラグを使用して、テンプレートファイルを必要とせずに新しい`.env`ファイルを最初から作成できます：

```bash
genenv -N
```

これにより、一般的な環境変数を含む一時的なテンプレートが作成され、インタラクティブに入力を促して値を設定できます。一時テンプレートには以下が含まれます：

- データベース設定（ホスト、ポート、名前、ユーザー、パスワード）
- API設定（キー、URL）
- アプリケーション設定（環境、デバッグモード、ログレベル、シークレットキー）

引数なしで`genenv`を実行し、プロンプトが表示されたときに最初から新しいファイルを作成することもできます。

### テンプレートメタデータ形式

フィールドの検証と説明を提供するために、テンプレートファイルにメタデータを追加できます。メタデータはフィールド定義の前のコメントで指定されます：

```txt
# @field_name [required] (type) 説明
KEY=${field_name}
```

例：

```txt
# @db_password [required] (string) データベースパスワード
DB_PASSWORD=${db_password}

# @db_port [optional] (int) データベースポート
DB_PORT=${db_port}

# @debug_mode [optional] (bool) デバッグモードを有効にする
DEBUG=${debug_mode}
```

サポートされているフィールドタイプ：

- `string`: テキスト値（デフォルト）
- `int`/`integer`: 整数値
- `bool`/`boolean`: 真偽値（true/false、yes/no、1/0）
- `float`/`double`: 浮動小数点値
- `url`: URL値（`http://`または`https://`で始まる必要があります）
- `email`: メールアドレス
- `ip`: IPv4またはIPv6アドレス（ネットワークから自動検出）
- `ipv4`: IPv4アドレスのみ（ネットワークから自動検出）
- `ipv6`: IPv6アドレスのみ（ネットワークから自動検出）

インタラクティブモードで実行すると、ツールはフィールドタイプに基づいて入力を検証し、フィールドの説明と共に適切なプロンプトを表示します。

### 既存の.envファイルとの比較

`-C/--compare`フラグを使用すると、genenvはテンプレートと既存の`.env`ファイルを比較し、新しいフィールドのみを追加します。これは、更新されたテンプレートから`.env`ファイルを新しいフィールドで更新する場合に便利です。

```bash
genenv -C .env.example
```

### 既存のフィールドのスキップ

`-S/--skip-existing`フラグを使用すると、genenvは`.env`ファイルに既に存在するフィールドをスキップします。これは、既に定義されているフィールドの既存の値を保持したい場合に便利です。

```bash
genenv -S .env.example
```

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
