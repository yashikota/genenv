# genenv

`genenv` は、`.evn.example` などのテンプレートファイルからプレースホルダーにランダムな値を入れた `.env` ファイルを生成するシンプルなCLIツールです  

*Read this in [English](README.md)*

## 例

`.env.example` ファイルが以下のような場合  

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

`genenv .env.example` を実行すると、以下のような `.env` ファイルが生成されます  

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

## インストール

[GitHub Releases](https://github.com/yashikota/genenv/releases)から最新のバイナリをダウンロードしてください  

もしくはGoでインストールしてください  

```bash
go install github.com/yashikota/genenv@latest
```

## 使い方

```bash
genenv .env.example
```

既存の`.env`ファイルが存在する場合、既存のフィールドの値は常に保持され、新しいフィールドに対してのみランダム値が生成されます  
置き換えて欲しくないプレースホルダーはバックスラッシュでエスケープします `\${not_a_placeholder}`  

### オプション

- `-f, --force`: 既存の値も含めてすべての値を再生成
  - `-y, --yes`: `--force` 使用時の確認プロンプトをスキップ
- `-o, --output`: 出力ファイルパスを指定（デフォルト: `.env`）
- `-l, --length`: 生成されるランダム値の長さ（デフォルト: 24）
- `-c, --charset`: 生成される値の文字セット
  - `alphanumeric`（デフォルト）: A-Z, a-z, 0-9
  - `alphabetic`: A-Z, a-z
  - `uppercase`: A-Z
  - `lowercase`: a-z
  - `numeric`: 0-9
- `-h, --help`: ヘルプ情報を表示
- `-v, --version`: バージョン情報を表示

### 長さと文字セットのカスタマイズ

`-l` と `-c` オプションを使用して、生成される値の長さと文字セットをカスタマイズできます  

```bash
# カスタムの長さ（32文字）で.envファイルを生成
genenv -l 32 .env.example

# カスタムの文字セット（数字のみ）で.envファイルを生成
genenv -c numeric .env.example

# カスタムの長さと文字セット（16文字の大文字）で.envファイルを生成
genenv -l 16 -c uppercase .env.example
```

### 再生成

デフォルトでは、`.env` ファイルの既存の値は保持されます。既存の値も含めてすべての値を再生成するには、`--force` フラグを使用します  
`--yes` もしくは `-y` を含めると確認をスキップします  

```bash
genenv --force --yes .env.example
```
