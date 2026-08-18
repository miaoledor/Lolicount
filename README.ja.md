# Lolicount

> 萌え系で着せ替え可能な SVG アクセスカウンター — 内蔵テーマを選ぶか自作テーマをアップロードして、リンクを貼るだけでカウント開始。

[中文](./README.md) · **日本語** · [English](./README.en.md)

Lolicount は萌え系の着せ替えアクセスカウンターで、SVG 画像として出力されます。
内蔵テーマをいくつか同梱しており、自作フレーム画像をアップロードして独自の
スタイルを作ることもできます。README やホームページにリンクを 1 行貼るだけで、
アクセスのたびに数字が +1 されます。

本プロジェクトは一度に 1 枚の画像を表示します。例えばフレームが
`0.png 1.png 2.png ... (n-1).png` のテーマの場合、アクセスごとに
`(count+1) % n` 番目のフレームを表示し、`count` を進めます。

## 特徴

- 🎀 **萌えテーマ** — 内蔵のロリ系フレーム画像、gif/png/webp 対応
- 🎨 **着せ替え** — 複数の内蔵テーマ、または自作フレームをアップロード可能
- 🖼️ **背景オーバーレイ** — カウンターを任意の背景画像に重ね合わせ
- 📊 **SVG 出力** — 鮮明なベクター、README に埋め込み可、JS 不要
- ⚡ **高性能** — Go + Fiber v3、メモリバッファ + バッチ書き込み SQLite
- 🛡️ **レート制限** — IP レベル + name レベルの二重レート制限、スパム防止
- 🚀 **シングルバイナリ** — フロントエンド + テーマを Go バイナリに埋め込み
- 🤝 **コミュニティ** — PR チャンネル(CI 自動検証)+ Web アップロードチャンネル

## クイックスタート

### Docker

```bash
docker run -d -p 9721:9721 \
  -v lolicount-data:/app/data \
  ghcr.io/miaoledor/lolicount:latest
```

`http://localhost:9721/@my-counter` を開いてください。カウントデータは
`lolicount-data` ボリューム内の SQLite ファイルに永続化されます。

### ソースから

```bash
git clone https://github.com/miaoledor/Lolicount.git
cd Lolicount
cp .env.example .env
go run ./cmd/server
```

### フロントエンド + バックエンドを同時に実行(開発モード)

ルートの `package.json` は `concurrently` を使い、バックエンド(Go :9721)と
フロントエンド(Nuxt :3721)を同時に起動します。macOS / Windows / Linux の
クロスプラットフォーム対応です:

```bash
pnpm install        # concurrently(ルート)とフロントエンド依存をインストール
pnpm dev            # フロントとバックを同時に起動
```

- バックエンド: http://127.0.0.1:9721
- フロントエンド: http://localhost:3721

個別に実行も可能: `pnpm dev:server`(バックのみ)または `pnpm dev:web`(フロントのみ)。

### 使い方

README やウェブページに埋め込む:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian)
```

背景オーバーレイ付き:

```markdown
![visitor](https://umi7.top/@my-counter?theme=lian&scale=2)
```

3 つの埋め込み形式(同じ URL):

```
1. SVG アドレス
   https://umi7.top/@my-counter?theme=lian

2. Img タグ
   <img src="https://umi7.top/@my-counter?theme=lian" alt="my-counter" />

3. Markdown
   ![my-counter](https://umi7.top/@my-counter?theme=lian)
```

## パラメータ

| パラメータ | 説明 | デフォルト |
|---|---|---|
| `theme` | テーマ名、または `random` | `lian` |
| `fsize` | カウンター文字サイズ(px) | `16` |
| `scale` | 画像表示倍率(統一された最長辺 400px 基準) | `1` |
| `number` | 指定数字のプレビュー(保存なし、+1 なし) | なし |
| `unshowf` | カウンター文字を非表示(`true`/`false`) | `false` |

> `scale` は画像サイズ、`fsize` は文字サイズを制御し、两者は独立です。
> `scale` を省略すると、すべてのテーマ画像は縦横比を保ったまま最長辺 400px に縮小されます。

## デフォルト設定

レンダリングのデフォルト値はすべて 1 つのファイルに集約されています:
`internal/theme/defaults.go`。デフォルトの挙動を変えるにはこのファイルを
編集するだけでよく、レンダリングロジックに触れる必要はありません。

| 定数 | 説明 | デフォルト |
|---|---|---|
| `DefaultTheme` | `?theme=` 省略時に使用されるテーマ | `lian` |
| `DefaultDisplaySize` | 統一最長辺目標(px)、`scale` 省略時に有効 | `400` |
| `DefaultFontSize` | `fsize` 省略時の文字サイズ | `16` |
| `MonoCharWidthFactor` | 等幅フォントの文字幅推定係数(文字サイズ比) | `0.6` |
| `DefaultFontFamily` | カウンター文字の CSS `font-family` | `monospace` |
| `DefaultFontColor` | カウンター文字の色 | `#333` |
| `TextGapBelowImage` | 画像下部と文字ベースラインの追加間隔(px) | `4` |

例 — デフォルト画像サイズを 600px、文字サイズを 20 にする:

```go
// internal/theme/defaults.go
const DefaultDisplaySize = 600
const DefaultFontSize   = 20
```

再ビルドで反映: `go build -o lolicount ./cmd/server && ./lolicount`

## API

| メソッド | パス | 説明 |
|---|---|---|
| GET | `/@:name` | カウント +1、SVG を返す |
| GET | `/get/@:name` | 同上(互換エイリアス) |
| GET | `/record/@:name` | JSON カウントを返す |
| GET | `/heart-beat` | ヘルスチェック |
| GET | `/api/themes` | テーマ一覧 |
| POST | `/api/themes` | テーマをアップロード |
| GET | `/api/backgrounds` | 背景一覧 |
| POST | `/api/backgrounds` | 背景をアップロード |

詳細は [docs/detail.md](./docs/detail.md) を参照してください。

## テーマの貢献

Lolicount は**フレーム式テーマ**を採用しています: 各テーマはディレクトリで、
中に複数のフレーム画像 `0.<ext> 1.<ext> ... n-1.<ext>` を含みます。アクセス
カウントは `(count+1) % n` でフレームを巡回します。拡張子 `gif` / `png` /
`webp` に対応し、フレームインデックスは 0 から連続している必要があります。

貢献方法は 2 つ:

**PR チャンネル** — リポジトリをフォークし、`assets/theme/<your-theme>/` に
フレーム(1 枚以上、インデックスは 0 から)と任意の `meta.json` を置いて PR を
開いてください。CI が自動実行されます:

- `cmd/check-theme` がディレクトリ名、フレーム完整性、形式とサイズを検証
- `scripts/validate-theme-meta.js` が `meta.json` のスキーマを検証
- `scripts/gen-themes-json.js` が `assets/themes.json` の同期を検証

**Web アップロード** — `/upload` ページでフレームをアップロード、即座に利用可能
(サーバー側で再エンコード)。

`meta.json` の例:

```json
{
  "name": "lian",
  "author": "yourname",
  "description": "Loli-style digit frames",
  "tags": ["cute", "anime"],
  "version": "1.0.0"
}
```

ローカル事前検証:

```bash
go run ./cmd/check-theme
node scripts/validate-theme-meta.js
node scripts/gen-themes-json.js
```

完全な貢献ガイドは [CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。

## CI/CD とデプロイ

| ワークフロー | トリガー | 目的 |
|---|---|---|
| `ci.yml` | push / PR | go vet + `go test -race` + check-theme + Nuxt build |
| `theme-check.yml` | PR が `assets/theme`, `assets/bg` を変更 | テーマ完全性 + meta.json + themes.json 同期 |
| `release.yml` | tag `v*` | Docker イメージ + Release バイナリをビルド |
| `rebuild-frontend.yml` | デフォルトブランチでテーマ変更 | SSG dist を再ビルドしてコミット |

**Docker**: `docker compose up -d`、`http://localhost:9721/@my-counter` を開く。
**Release**: `git tag v0.1.0 && git push --tags` でタグを付ける、CI がイメージをビルドして Release を公開します。

## 技術スタック

- **バックエンド**: Go 1.23+ / Fiber v3 / SQLite(`modernc.org/sqlite`、純 Go、CGO 不要)
- **フロントエンド**: Nuxt 3 SSG / UnoCSS / GSAP
- **ストレージ**: リクエスト → メモリバッファ → バッチ書き込み → SQLite
- **デプロイ**: シングルバイナリ(embed.FS がテーマ + フロント dist を同梱)

## 謝辞

- [kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum) — ロリキャラのレイヤー重ね合わせ手法
- [Moe-Counter](https://github.com/journey-ad/Moe-Counter) — 本プロジェクトの元となった Moe-Counter

## スポンサー

Lolicount がお役に立ちましたら、[作者をスポンサー](https://github.com/sponsors/miaoledor)してください 🧋
