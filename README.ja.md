![Lolicount](docs/png/lolicount-s.png)

[English](./README.md) · [中文](./README.zh.md) · **日本語**

GitHub README、ブログ、個人サイトに貼れる、キャラクターテーマ付きアクセスカウンター。キャラクターを選び、埋め込みコードをコピーするだけで使えます。静的テーマは、リクエストごとに画像をランダムに選んだり、立ち絵のレイヤーを組み合わせたりできます。

## 使い方

1. [lolicount.top](https://lolicount.top) を開きます。
2. カウンター名とテーマを選び、生成ボタンを押します。
3. Markdown または HTML コードをコピーして、ページに貼り付けます。

以下の `my-counter` を自分の名前に置き換えてください。同じ名前のカウンターは、複数のページに貼ってもカウントを共有します。

**GitHub README / Markdown**

```markdown
![アクセス数](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6)
```

**Web サイト / HTML**

```html
<img src="https://lolicount.top/@my-counter?theme=lian-ren&amp;fsize=16&amp;scale=0.6" alt="アクセス数" />
```

**画像 URL**：[サンプルを開く](https://lolicount.top/@my-counter?theme=lian-ren&fsize=16&scale=0.6)。画像 URL を指定できる場所に、そのアドレスを貼り付けることもできます。

GitHub README には SVG テーマを選んでください。E-mote、Spine、Live2D のアニメーションテーマは、サイトで生成される Web ウィジェット用コードを使い、JavaScript が必要です。

## カスタマイズ

URL パラメーターを変更するか、サイトで調整しながらプレビューできます。

| パラメーター | 用途 | 例 |
|---|---|---|
| `theme` | キャラクターテーマ。`random` で SVG テーマをランダムに選択 | `theme=lian-ren` |
| `scale` | 画像の拡大率。`0.1`～`4`、既定値 `1` | `scale=0.6` |
| `fsize` | カウント文字のサイズ（px）。`0`～`500`、`0` で既定サイズ | `fsize=16` |
| `number` | `1`～`999999` の固定値を表示し、加算しない。通常の計数時は省略 | `number=12345` |

計数せずに試すには、名前を [`demo`](https://lolicount.top/@demo?theme=lian-ren&fsize=16&scale=0.6) に変更します。既定では `0123456789` を表示し、`number` を指定するとその値を表示します。

フォントや文字位置の設定は[パラメーター一覧](./docs/projectDesign.md#接口调用文档)、困ったときは [FAQ](./docs/faq.md) をご覧ください。

## セルフホスト

Docker で起動します。

```sh
docker run -d --name lolicount --restart unless-stopped -p 9721:9721 -v lolicount-data:/app/data ghcr.io/miaoledor/lolicount:latest
```

[localhost:9721](http://localhost:9721) を開くと、カウンターを生成できます。データは `lolicount-data` ボリュームに保存されます。単一インスタンスで運用し、公開ドメインを使う場合は `BASE_URL` に `https://counter.example.com` のような公開 URL を設定してください。

Docker Compose、環境設定、バイナリのビルドは[デプロイガイド](./docs/deployment.md)をご覧ください。

## ローカル開発

Go 1.25+、Node.js 22+、pnpm 11 を用意し、リポジトリのルートで実行します。

```sh
pnpm install
pnpm dev
```

フロントエンドは [localhost:3721](http://localhost:3721)、バックエンドは [127.0.0.1:9721](http://127.0.0.1:9721) で動作します。

## 貢献と支援

- [貢献ガイド](./CONTRIBUTING.md) · [テーマの追加](./docs/contributing-themes.md) · [コードの貢献](./docs/contributing-code.md)
- [ドキュメント](./docs/index.md) · [アーキテクチャ](./docs/architecture.md) · [ロードマップ](./docs/TODOlist.md)
- 関連ツール：[Lolicount Editor](https://github.com/miaoledor/Lolicount-Editor) · [Yuzusoft Tachie skill](https://github.com/miaoledor/yuzusoft-tachie-skill)
- [作者にミルクティーを一杯](https://github.com/sponsors/miaoledor) 🧋

[kun-galgame-forum](https://github.com/KunMoe/kun-galgame-forum) と [Moe-Counter](https://github.com/journey-ad/Moe-Counter) に感謝します。

[AGPL-3.0](./LICENSE) ライセンスで公開しています。

![Lolicount](https://lolicount.top/@lolicount?theme=lian-ren&fsize=16&scale=1&unshowf=true)
