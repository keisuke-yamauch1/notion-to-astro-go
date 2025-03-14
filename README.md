# Notion to Astro 変換ツール

このツールは、Notionデータベースから記事を取得し、Astroテンプレート形式（Markdownファイル + YAMLフロントマター）に変換するためのGo言語で書かれたユーティリティです。

## 前提条件

- Go 1.24以上
- Notion APIトークン
- Notionブログデータベース ID（ブログデータベースを使用する場合）
- Notion日記データベース ID（日記データベースを使用する場合）

## インストール

```bash
git clone https://github.com/yourusername/notion-to-astro-go.git
cd notion-to-astro-go
go mod download
```

## 使い方

### 環境変数の設定

環境変数は2つの方法で設定できます：

#### 1. .envファイルを使用する方法（推奨）

1. `.env.example`ファイルを`.env`にコピーします：

```bash
cp .env.example .env
```

2. `.env`ファイルを編集して、必要な情報を入力します：

```
NOTION_API_TOKEN=your_notion_api_token
NOTION_BLOG_DATABASE_ID=your_notion_blog_database_id
NOTION_DIARY_DATABASE_ID=your_notion_diary_database_id
BLOG_OUTPUT_DIR=./content/blog  # ブログ記事の出力先ディレクトリ
DIARY_OUTPUT_DIR=./content/diary  # 日記エントリの出力先ディレクトリ
IMAGES_DIR=./public/images  # Notionから取得した画像の保存先ディレクトリ
```

#### 2. 直接環境変数を設定する方法

```bash
export NOTION_API_TOKEN="your_notion_api_token"
export NOTION_BLOG_DATABASE_ID="your_notion_blog_database_id"
export NOTION_DIARY_DATABASE_ID="your_notion_diary_database_id"
export BLOG_OUTPUT_DIR="./content/blog"  # ブログ記事の出力先ディレクトリ
export DIARY_OUTPUT_DIR="./content/diary"  # 日記エントリの出力先ディレクトリ
export IMAGES_DIR="./public/images"  # Notionから取得した画像の保存先ディレクトリ
```

### 実行

デフォルトでは、すべてのデータベースタイプ（ブログと日記）が処理されます：

```bash
go run main.go
```

特定のデータベースタイプを指定するには、`-type`フラグを使用します：

```bash
# すべてのデータベースタイプを処理する場合（デフォルト）
go run main.go -type all

# ブログデータベースのみを処理する場合
go run main.go -type blog

# 日記データベースのみを処理する場合
go run main.go -type diary
```

## 機能

- Notionデータベースから記事を取得
- 記事のプロパティからフロントマターを抽出（タイトル、説明、タグ、日付、下書きステータスなど）
- Notionのブロックコンテンツをマークダウンに変換
- 変換された記事をAstro互換のマークダウンファイルとして保存
- ブログ記事の場合、最初の70文字を自動的に説明文として使用
- 日記エントリの場合、説明文と天気情報を抽出
- 空行の処理：段落間の単一の空行を削除し、複数の連続した空行がある場合は1つだけ保持
- 画像の処理：Notionの画像を自動的にダウンロードし、Astroプロジェクトの指定されたディレクトリに保存して、マークダウン内の参照を更新

## フィルタリング

このツールは、Notionデータベースから記事を取得する際に以下のフィルタを適用します：

- `published` プロパティが `false` の記事
- `done` プロパティが `true` の記事

これにより、公開準備が完了しているが、まだ公開されていない記事のみが処理されます。

## サポートされているNotionブロック

- 段落
- 見出し（H1、H2、H3）
- 箇条書きリスト
- 番号付きリスト
- ToDo
- コードブロック（言語シンタックスハイライト付き）
- 引用
- 区切り線
- 画像（外部URLと内部ファイル）
- リンク（リッチテキスト内のリンク）

## Notionデータベースの設定

このツールは以下のプロパティを持つNotionデータベースを想定しています：

### 共通プロパティ
- `title`/`Title`/`Name`: 記事のタイトル（必須）
- `tags`/`Tags`: 記事のタグ（マルチセレクト、オプション）
- `published`: 公開ステータス（チェックボックス、オプション）
- `done`: 完了ステータス（チェックボックス、オプション）
- `ID`/`id`: 記事のID（オプション、指定されていない場合はNotionのページIDが使用されます）

### ブログデータベース固有のプロパティ
- 説明文は記事の最初の70文字から自動的に生成されます

### 日記データベース固有のプロパティ
- `description`/`Description`: 日記の説明（リッチテキスト、オプション）
- `weather`: 天気情報（リッチテキスト、オプション）

これらのプロパティが存在しない場合、デフォルト値または空の値が使用されます。

## 出力形式

ブログ記事は `BLOG_OUTPUT_DIR` で指定されたディレクトリに保存され、日記エントリは `DIARY_OUTPUT_DIR` で指定されたディレクトリに保存されます。

### ブログ記事

```markdown
---
id: article-id
title: 記事のタイトル
description: 記事の説明（最初の70文字から自動生成）
date: 2023-01-01
tags: ["タグ1", "タグ2"]
draft: true
---

# 記事の内容

これは記事の本文です。
```

### 日記エントリ

```markdown
---
id: diary-id
title: 日記のタイトル
description: 日記の説明
date: 2023-01-01
tags: ["タグ1", "タグ2"]
weather: 晴れ
draft: true
---

# 日記の内容

これは日記の本文です。
```

ファイル名は記事のタイトルに基づいて生成され、スペースやその他の特殊文字はハイフンに置き換えられます。

## 空行の処理

このツールは、以下のルールに従って空行を処理します：

- 段落間の単一の空行は削除されます
- 複数の連続した空行がある場合は、1つだけ保持されます
- フロントマターの後の最初の空行は保持されます

これにより、出力されるマークダウンファイルは一貫した形式になります。

## トラブルシューティング

- `NOTION_API_TOKEN environment variable is required`: NOTION_API_TOKEN環境変数が設定されていません
- `NOTION_BLOG_DATABASE_ID environment variable is required for blog database`: ブログデータベースを処理する場合、NOTION_BLOG_DATABASE_ID環境変数が設定されていません
- `NOTION_DIARY_DATABASE_ID environment variable is required for diary database`: 日記データベースを処理する場合、NOTION_DIARY_DATABASE_ID環境変数が設定されていません
- `Invalid database type: X. Must be 'blog' or 'diary'`: 無効なデータベースタイプが指定されました。'blog'または'diary'を指定してください
- `Failed to create blog output directory`: ブログ記事の出力ディレクトリの作成に失敗しました
- `Failed to create diary output directory`: 日記エントリの出力ディレクトリの作成に失敗しました
- `Failed to create images directory`: 画像の出力ディレクトリの作成に失敗しました
- `Failed to get database`: Notionデータベースの取得に失敗しました
- `Failed to query database`: Notionデータベースのクエリに失敗しました
- `Failed to convert article`: 記事のAstroテンプレートへの変換に失敗しました
- `Failed to write article to file`: 記事のファイルへの書き込みに失敗しました
- `Failed to download image`: 画像のダウンロードに失敗しました。この場合、元のNotionの画像URLが使用されます
- `Failed to create output file`: 画像ファイルの作成に失敗しました
- `Failed to save image`: 画像の保存に失敗しました
