![miaoledor](docs/png/githubSocialPreview.png)
[English](./README.md) · [中文](./README.zh.md) · **日本語**

### 外部画像ソースをサポートするホームページなどで、お気に入りのキャラクターを表示しよう！

萌え系で着せ替え可能なアクセスカウンター。SVG 画像として出力されます。内蔵テーマをいくつか同梱しており、自作の数字画像や背景をアップロードして独自のスタイルを作ることもできます。README やホームページにリンクを 1 行貼るだけ！


表示されるキャラクターはランダムなフレーム選択とランダムなレイヤー合成に対応し、galgame のキャラクター立絵のような`動的合成`に対応しています。

## 少ない素材で、万の変化 — 最小のストレージで最も豊かな変化を

多レイヤーテーマはキャラクターを表情、目、口、顔などの独立レイヤーに分割し、リクエストごとにランダムに合成します。`lian-ren` テーマを例にすると、わずか **71 枚**の画像（lass ×8 + brow ×18 + eye ×18 + mouth ×20 + face ×6）で **311,040 通り**の立絵の組み合わせを生成します——リロードのたびに新しい姿が現れます。

## 高速読み込み

Fiber v3 は [Fasthttp](https://gofiber.io/) を基盤にしており、Fiber チームはこれを Go 最速の HTTP エンジンと紹介しています。[Fiber 公式の TechEmpower ベンチマーク](https://docs.gofiber.io/extra/benchmarks/)では、Plaintext スループットが Fiber v3 約 **1,198.8 万応答/秒**、Express 約 **120.5 万**、JSON シリアライズが Fiber v3 約 **236.3 万応答/秒**、Express 約 **94.97 万** です。

Lolicount では、この軽量なリクエスト経路に埋め込みテーマ素材と Nuxt SSG フロントエンドを組み合わせます。カウンター SVG はプロセス内で合成されるため、テーマ取得の追加リクエストが不要で、すばやく応答します。

## クイックスタート

### 直接使用
https://lolicount.top をご覧ください。

### テスト開発実行

ルートの `package.json` は `concurrently` を使い、バックエンド(Go :9721)とフロントエンド(Nuxt :3721)を同時に起動します。macOS / Windows / Linux のクロスプラットフォーム対応です:

```bash
pnpm install        # concurrently(ルート)とフロントエンド依存をインストール
pnpm dev            # フロントとバックを同時に起動
```

個別に実行も可能: `pnpm dev:server`(バックのみ)または `pnpm dev:web`(フロントのみ)。

### サーバーデプロイ

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

`http://localhost:9721/@my-counter` にアクセスしてください。カウントデータは `lolicount-data` ボリューム内の SQLite ファイルに永続化されます。

CI/CD は GitHub Actions で構成:プッシュ時に `go vet` + テストが自動実行され、`v*` タグのプッシュでフロントエンドのビルド、静的バイナリのコンパイル、Docker イメージの ghcr.io へのプッシュが自動的に行われます。

## 貢献

私たちは本当にあなたの助けを必要としています！

機能の豊富化であれ、テーマの追加であれ、あなたの参加が必要です。
貢献の`詳細`はこちら:
| ドキュメント | 内容 |
|---|---|
| [CONTRIBUTING.md](./CONTRIBUTING.md) | 貢献の概要 |
| [docs/contributing-themes.md](./docs/contributing-themes.md) | テーマ貢献ガイド |
| [docs/contributing-code.md](./docs/contributing-code.md) | 機能貢献ガイド |

## 謝辞

- [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum)
- [Moe-Counter](https://github.com/journey-ad/Moe-Counter)

## スポンサー

このプロジェクトが好きですか?Lolicount がお役に立ちましたら、[作者にタピオカを一杯おごって](https://github.com/sponsors/miaoledor)ください 🧋

## 技術スタック

**バックエンド**: Go 1.25+ / Fiber v3 / SQLite
**パフォーマンス**: Fiber v3 による高速なリクエスト処理で、埋め込み流量でもカウンター SVG をすばやく応答します。
**フロントエンド**: Vue(Nuxt 4 SSG)/ UnoCSS / GSAP
**ストレージ**: リクエスト → メモリバッファ → バッチ書き込み → SQLite
**デプロイ**: シングルバイナリ(embed.FS がテーマ + フロント dist を同梱)
技術的な詳細は以下のドキュメントでご覧いただけます:
| ドキュメント | 内容 |
|---|---|
| [docs/architecture.md](./docs/architecture.md) | アーキテクチャ、プロジェクト構造、技術選定 |
| [docs/deployment.md](./docs/deployment.md) | 使用とデプロイ(Win/Mac/Linux) |
| [docs/projectDesign.md](./docs/projectDesign.md) | プロジェクト設計とインターフェース契約 |
| [docs/TODOlist.md](./docs/TODOlist.md) | マイルストーンとタスク状態 |

## ライセンス

本プロジェクトは [AGPL-3.0](./LICENSE) ライセンスの下で公開されています。

本项目基于 [AGPL-3.0](./LICENSE) 协议开源。

![lolicount](https://lolicount.top/@lolicount?theme=lian-ren&fsize=16&scale=1&unshowf=true)
