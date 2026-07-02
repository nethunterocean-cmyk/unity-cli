# unity-cli

[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [Français](README.fr.md)

> コマンドラインから Unity Editor を操作する。AIエージェントのために作ったけど、何でも使えるで。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**サーバー起動不要。設定ファイル不要。プロセス管理不要。コマンド叩くだけ。**

> **🔒 セキュリティフォーク:** 自動更新とバージョンチェック機能はセキュリティ上削除されました。
> 自動更新が欲しい人はオリジナル → [youngwoocho02/unity-cli](https://github.com/youngwoocho02/unity-cli) へどうぞ。

## ビルド (CLI)

```bash
git clone https://github.com/nethunterocean-cmyk/unity-cli
cd unity-cli
go build -o unity-cli .
sudo mv unity-cli /usr/local/bin/
```

[Go](https://go.dev/dl/) 1.24+ が必要。他の依存関係なし。

## Unity セットアップ

`UnityFiles/` フォルダを Unity プロジェクトの `Assets/` にコピー:

```bash
cp -r UnityFiles /path/to/YourUnityProject/Assets/
```

コネクタは Unity 起動時に自動で動き出す。設定不要。ほな、ええ感じに使ってな。
