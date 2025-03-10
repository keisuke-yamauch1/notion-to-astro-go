# Notion to Astro 変換ツール

このツールは、Notionデータベースから記事を取得し、Astroテンプレート形式（Markdownファイル + YAMLフロントマター）に変換するためのGo言語で書かれたユーティリティです。

## 前提条件

- Go 1.24以上
- Notion APIトークン
- NotionデータベースID

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
NOTION_DATABASE_ID=your_notion_database_id
OUTPUT_DIR=./content  # 出力先ディレクトリ（デフォルトは ./content）
```

#### 2. 直接環境変数を設定する方法

```bash
export NOTION_API_TOKEN="your_notion_api_token"
export NOTION_DATABASE_ID="your_notion_database_id"
export OUTPUT_DIR="./content"  # 出力先ディレクトリ（デフォルトは ./content）
```

### 実行

```bash
go run main.go
```

## 機能

- Notionデータベースから記事を取得
- 記事のプロパティからフロントマターを抽出（タイトル、説明、タグ、公開日、更新日、下書きステータスなど）
- Notionのブロックコンテンツをマークダウンに変換
- 変換された記事をAstro互換のマークダウンファイルとして保存

## サポートされているNotionブロック

- 段落
- 見出し（H1、H2、H3）
- 箇条書きリスト
- 番号付きリスト
- ToDo
- コードブロック
- 引用
- 区切り線
- 画像

## Notionデータベースの設定

このツールは以下のプロパティを持つNotionデータベースを想定しています：

- `title`/`Title`/`Name`: 記事のタイトル（必須）
- `description`/`Description`: 記事の説明（オプション）
- `tags`/`Tags`: 記事のタグ（マルチセレクト、オプション）
- `published`/`Published`/`PublishedAt`: 公開日（日付、オプション）
- `draft`/`Draft`: 下書きステータス（チェックボックス、オプション）

これらのプロパティが存在しない場合、デフォルト値または空の値が使用されます。

## 出力形式

各記事は以下の形式のマークダウンファイルとして保存されます：

```markdown
---
title: 記事のタイトル
description: 記事の説明
publishedAt: 2023-01-01
updatedAt: 2023-01-02
tags:
  - タグ1
  - タグ2
draft: false
---

# 記事の内容

これは記事の本文です。
```

ファイル名は記事のタイトルに基づいて生成され、スペースやその他の特殊文字はハイフンに置き換えられます。

## トラブルシューティング

- `NOTION_API_TOKEN environment variable is required`: NOTION_API_TOKEN環境変数が設定されていません
- `NOTION_DATABASE_ID environment variable is required`: NOTION_DATABASE_ID環境変数が設定されていません
- `Failed to create output directory`: 出力ディレクトリの作成に失敗しました
- `Failed to fetch articles`: Notionデータベースからの記事の取得に失敗しました
- `Failed to convert article`: 記事のAstroテンプレートへの変換に失敗しました
- `Failed to write article to file`: 記事のファイルへの書き込みに失敗しました
