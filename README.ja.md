# unity-cli

[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [Français](README.fr.md)

> コマンドラインから Unity Editor を操作する。AIエージェントのために作ったけど、何でも使える。

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

コネクタは Unity 起動時に自動で動き出す。設定不要。

### 推奨: Editor スロットリングを無効にする

デフォルトでは、Unity はウィンドウがフォーカスを失うとエディタの更新をスロットリングする。CLI コマンドは Unity メインスレッドで処理されるため、応答が遅れる可能性がある。

**Edit → Preferences → General → Interaction Mode** を **No Throttling** に設定せよ。

コネクタも CLI リクエストごとに PlayerLoop 更新を要求する。それでもバックグラウンドでの応答性を確実にするには No Throttling が推奨される。

## クイックスタート

```bash
# Unity 接続確認
unity-cli status

# プレイモードに入って待つ
unity-cli editor play --wait

# Unity 内で C# コードを実行
unity-cli exec "return Application.dataPath;"

# コンソールログを読む
unity-cli console --type error,warning,log
```

## 仕組み

```
端末                                   Unity Editor
────                                   ────────────
$ unity-cli editor play --wait
    │
    ├─ ~/.unity-cli/instances/*.json をスキャン
    │  → このプロジェクトの Unity インスタンスを選択
    │
    ├─ 選択した Unity listener にコマンドを送信
    │  { "command": "manage_editor",
    │    "params": { "action": "play",
    │                "wait_for_completion": true }}
    │                                      │
    │                                  HttpServer 受信
    │                                      │
    │                                  CommandRouter ディスパッチ
    │                                      │
    │                                  ManageEditor.HandleCommand()
    │                                  → EditorApplication.isPlaying = true
    │                                  → PlayModeStateChange を待機
    │                                      │
    ├─ JSON 応答を受信  ←─────────────────┘
    │  { "success": true,
    │    "message": "Entered play mode (confirmed)." }
    │
    └─ 出力: Entered play mode (confirmed).
```

Unity コネクタの動作:
1. Editor 起動時にローカル HTTP listener を開く
2. `~/.unity-cli/instances/` にプロジェクト別のインスタンスファイルを書き込み、CLI が接続先を把握できるようにする
3. 0.5秒ごとに現在の状態をインスタンスファイルに更新 (heartbeat)
4. リクエストごとにリフレクションで `[UnityCliTool]` クラスを発見
5. 受信したコマンドをメインスレッドの該当ハンドラにルーティング
6. ドメインリロード（スクリプト再コンパイル）後も維持される

コンパイルやリロードの直前に、状態 (`compiling`、`reloading`) をインスタンスファイルに記録する。メインスレッドが停止すると timestamp の更新が止まり、CLI は新しい timestamp が書き込まれるまで待機してからコマンドを送信する。

## 内蔵コマンド

| コマンド | 説明 |
|---------|------|
| `editor` | Unity Editor の play/stop/pause/refresh を制御 |
| `console` | コンソールログの読み取り、フィルタリング、消去 |
| `exec` | Unity 内で任意の C# コードを実行 |
| `test` | EditMode/PlayMode テストを実行 |
| `menu` | Unity メニュー項目をパスで実行 |
| `reserialize` | Unity シリアライザ経由でアセットを再シリアライズ |
| `screenshot` | Scene/Game ビューを PNG でキャプチャ |
| `profiler` | プロファイラ階層の読み取り、録画制御 |
| `list` | 利用可能な全ツールとパラメータスキーマを表示 |
| `status` | Unity Editor の接続状態を表示 |
| `update` | 自動更新 — セキュリティ上無効化 |

### Editor 制御

```bash
# プレイモードに入る
unity-cli editor play

# プレイモードに入り、完全にロードされるまで待つ
unity-cli editor play --wait

# プレイモードを終了
unity-cli editor stop

# 一時停止トグル (プレイモード中のみ)
unity-cli editor pause

# アセットをリフレッシュ (プレイモード中は --force が必要)
unity-cli editor refresh

# リフレッシュ + スクリプトコンパイル
unity-cli editor refresh --compile

# プレイモード中でも強制リフレッシュ
unity-cli editor refresh --force
```

### コンソールログ

```bash
# エラーと警告ログを読む (デフォルト)
unity-cli console

# 最新20件の全タイプログ
unity-cli console --lines 20 --filter error,warning,log

# エラーのみ
unity-cli console --type error

# スタックトレース付き (user: ユーザーコードのみ, full: 生データ)
unity-cli console --stacktrace user

# コンソールを消去
unity-cli console --clear
```

### C# コード実行

Unity Editor ランタイムで任意の C# コードを実行できる。最も強力なコマンド。UnityEngine、UnityEditor、ECS、ロード済みの全アセンブリにフルアクセス可能。使い捨てのクエリや変更のためにカスタムツールを書く必要はない。

`return` で出力を取得。主要な名前空間はデフォルトで使用可能。プロジェクト固有の型（例: `Unity.Entities`）のみ `--usings` で追加。カンマ区切りで複数指定可能、繰り返し指定も可。

```bash
unity-cli exec "return Application.dataPath;"
unity-cli exec "return EditorSceneManager.GetActiveScene().name;"
unity-cli exec "return World.All.Count;" --usings Unity.Entities
unity-cli exec "return World.All.Count;" --usings Unity.Entities --usings Unity.Mathematics

# stdin からのパイプでシェルエスケープ問題を回避
echo 'Debug.Log("hello"); return null;' | unity-cli exec
echo 'var go = new GameObject("Marker"); go.tag = "EditorOnly"; return go.name;' | unity-cli exec
```

`exec` はデフォルトで非同期、コルーチン、Unity 遅延コールバックをブロックする。意図的に遅延動作を許可する場合のみ `--allow-async` を使用。

`exec` は実際の C# をコンパイルして実行するため、カスタムツールができることはすべて可能 — ECS エンティティの調査、アセットの変更、内部 API の呼び出し、エディタユーティリティの実行。AI エージェントにとっては **ツールコードを一行も書かずに Unity ランタイム全体にゼロフリクションでアクセス**できることを意味する。stdin パイプを使えば複雑なコードでもシェルエスケープの問題が起きない。

### メニュー項目

```bash
# Unity メニュー項目をパスで実行
unity-cli menu "File/Save Project"
unity-cli menu "Assets/Refresh"
unity-cli menu "Window/General/Console"
```

安全のため `File/Quit` はブロックされている。

### アセット再シリアライズ

AI エージェント（と人間）は Unity アセットファイル（`.prefab`、`.unity`、`.asset`、`.mat`）をプレーンテキスト YAML として直接編集できる。しかし Unity の YAML シリアライザは厳密で、フィールド欠落やインデントミス、古い `fileID` 一つでアセットが静かに壊れる。

`reserialize` はこれを解決する。テキスト編集後に実行すると、アセットをメモリにロードし、Unity 自身のシリアライザで書き出し直す。インスペクタで編集したのと同じ、クリーンで有効な YAML ファイルが生成される。

```bash
# プロジェクト全体を再シリアライズ (引数なし)
unity-cli reserialize

# プレハブの Transform 値をテキスト編集した後
unity-cli reserialize Assets/Prefabs/Player.prefab

# 複数シーンを一括編集した後
unity-cli reserialize Assets/Scenes/Main.unity Assets/Scenes/Lobby.unity

# マテリアルプロパティ変更後
unity-cli reserialize Assets/Materials/Character.mat
```

これがテキストベースのアセット編集を安全にする仕組み。これがないと、YAML フィールドを一箇所間違えただけでランタイムまで発覚しない prefab 破損が発生する。これがあれば **AI エージェントはどんな Unity アセットでもテキストで自信を持って編集**できる — prefab へのコンポーネント追加、シーン階層の変更、マテリアルプロパティの調整 — 結果が正しくロードされることを保証しながら。

### プロファイラ

```bash
# プロファイラ階層を読む (最終フレーム、最上位)
unity-cli profiler hierarchy

# 再帰的ドリルダウン
unity-cli profiler hierarchy --depth 3

# 名前でルート指定 (部分一致) — 特定システムにフォーカス
unity-cli profiler hierarchy --root SimulationSystem --depth 3

# 特定のアイテム ID でドリルダウン
unity-cli profiler hierarchy --parent 4 --depth 2

# 最新30フレームの平均
unity-cli profiler hierarchy --frames 30 --min 0.5

# 特定フレーム範囲の平均
unity-cli profiler hierarchy --from 100 --to 200

# フィルタとソート
unity-cli profiler hierarchy --min 0.5 --sort self --max 10

# プロファイラ録画 ON/OFF
unity-cli profiler enable
unity-cli profiler disable

# プロファイラ状態確認
unity-cli profiler status

# キャプチャしたフレームを消去
unity-cli profiler clear
```

### テスト実行

Unity Test Framework 経由で EditMode/PlayMode テストを実行。

```bash
# EditMode テスト実行 (デフォルト)
unity-cli test

# PlayMode テスト実行
unity-cli test --mode PlayMode

# テスト名でフィルタ (部分一致)
unity-cli test --filter MyTestClass
```

Unity Test Framework パッケージが必要。PlayMode テストはドメインリロードを引き起こすが、CLI が結果ファイルをポーリングする。

### ツール一覧

```bash
# 利用可能な全ツール (内蔵 + プロジェクトカスタム) をパラメータスキーマ付きで表示
unity-cli list
```

### カスタムツール

```bash
# カスタムツールを名前で直接呼び出し
unity-cli my_custom_tool

# パラメータ付きで呼び出し
unity-cli my_custom_tool --params '{"key": "value"}'
```

### 状態確認

```bash
# Unity Editor の状態を表示
unity-cli status
# 出力: Unity: ready
#   Project: /path/to/project
#   Version: 6000.1.0f1
#   PID:     12345
```

コマンド送信前に CLI が自動的に Unity の状態を確認する。Unity がビジー（コンパイル中、リロード中）なら応答可能になるまで待機する。

## グローバルオプション

| フラグ | 説明 | デフォルト |
|-------|------|-----------|
| `--project <path>` | プロジェクトパスで Unity インスタンスを選択 | auto |
| `--timeout <ms>` | HTTP リクエストタイムアウト | 120000 |
| `--ignore-version-mismatch` | CLI/connector バージョンチェックをスキップ | false |

```bash
# 複数インスタンスからプロジェクトパスで選択
unity-cli --project MyGame editor stop

# CLI と connector のバージョンが異なっても実行
unity-cli --ignore-version-mismatch status
```

すべてのコマンドで `--help` を付けると詳細な使い方が表示される:

```bash
unity-cli editor --help
unity-cli exec --help
unity-cli profiler --help
```

## カスタムツールの作成

Editor アセンブリに `[UnityCliTool]` 属性を持つ static クラスを作成する。ドメインリロード時に自動検出される。

```csharp
using UnityCliConnector;
using Newtonsoft.Json.Linq;
using UnityEngine;

[UnityCliTool(Name = "spawn", Description = "指定位置に敵をスポーン", Group = "gameplay")]
public static class SpawnEnemy
{
    public class Parameters
    {
        [ToolParameter("X ワールド座標", Required = true)]
        public float X { get; set; }

        [ToolParameter("Y ワールド座標", Required = true)]
        public float Y { get; set; }

        [ToolParameter("Z ワールド座標", Required = true)]
        public float Z { get; set; }

        [ToolParameter("Resources フォルダ内のプレハブ名", DefaultValue = "Enemy")]
        public string Prefab { get; set; }
    }

    public static object HandleCommand(JObject parameters)
    {
        var p = new ToolParams(parameters);
        float x = p.GetFloat("x", 0);
        float y = p.GetFloat("y", 0);
        float z = p.GetFloat("z", 0);
        string prefabName = p.Get("prefab", "Enemy");

        var prefab = Resources.Load<GameObject>(prefabName);
        var instance = Object.Instantiate(prefab, new Vector3(x, y, z), Quaternion.identity);

        return new SuccessResponse("Enemy spawned", new
        {
            name = instance.name,
            position = new { x, y, z }
        });
    }
}
```

フラグまたは JSON で呼び出し:

```bash
unity-cli spawn --x 1 --y 0 --z 5 --prefab Goblin
unity-cli spawn --params '{"x":1,"y":0,"z":5,"prefab":"Goblin"}'
```

**重要ポイント:**

- **名前**: `Name` がない場合はクラス名から自動生成 (`SpawnEnemy` → `spawn_enemy`、`UITree` → `ui_tree`)。`Name = "spawn"` なら `unity-cli spawn` で呼び出し可能。
- **Parameters クラス**: 任意だが推奨。`unity-cli list` でパラメータ名、型、説明、必須フラグを公開 — AI アシスタントがソースコードなしでツールの使い方を把握できる。
- **ToolParams**: `p.Get()`、`p.GetInt()`、`p.GetFloat()`、`p.GetBool()`、`p.GetRaw()` で一貫したパラメータ読み取り。
- **検出**: `unity-cli list` で内蔵ツール (`group: "built-in"`) が先に、接続されたプロジェクトのカスタムツール (`group: "custom"`) が後に表示される。

**属性リファレンス:**

| 属性 | プロパティ | 説明 |
|------|-----------|------|
| `[UnityCliTool]` | `Name` | コマンド名のオーバーライド (デフォルト: クラス名 → snake_case) |
| | `Description` | `list` に表示されるツールの説明 |
| | `Group` | 分類用のグループ名 |
| `[ToolParameter]` | `Description` | パラメータの説明 (コンストラクタ引数) |
| | `Required` | 必須かどうか (デフォルト: `false`) |
| | `Name` | パラメータ名のオーバーライド |
| | `DefaultValue` | デフォルト値のヒント |

### ルール

- クラスは `static` であること
- `public static object HandleCommand(JObject parameters)` または `async Task<object>` 版が必要
- `SuccessResponse(message, data)` または `ErrorResponse(message)` を返すこと
- `Parameters` 入れ子クラスに `[ToolParameter]` 属性を追加すると自動文書化される
- クラス名は自動的に snake_case のコマンド名に変換される
- `[UnityCliTool(Name = "my_name")]` で名前を上書き可能
- Unity メインスレッドで実行されるため、すべての Unity API を安全に呼び出せる
- Editor 起動時およびスクリプト再コンパイル後に自動検出される
- 重複するツール名は検出されてエラーログに記録される — 最初に見つかったハンドラのみ使用される

## 複数の Unity インスタンス

複数の Unity Editor が開いている場合、それぞれがプロジェクトパスを登録する:

```bash
# 実行中の全インスタンスを表示
unity-cli status

# プロジェクトパスで選択
unity-cli --project MyGame editor play

# デフォルト: カレントディレクトリの Unity プロジェクト、または唯一のアクティブインスタンスを使用
unity-cli editor play
```

## MCP との比較

| | MCP | unity-cli |
|---|-----|-----------|
| **インストール** | Python + uv + FastMCP + config JSON | 単一バイナリ |
| **依存関係** | Python ランタイム、WebSocket リレー | なし |
| **プロトコル** | JSON-RPC 2.0 over stdio + WebSocket | 直接 HTTP POST |
| **セットアップ** | MCP 設定生成、AI ツール再起動 | Unity パッケージ追加、完了 |
| **再接続** | 複雑なドメインリロード再接続ロジック | リクエストごとにステートレス |
| **互換性** | MCP 互換クライアントのみ | シェルがあるすべてのもの |
| **カスタムツール** | 同じ `[Attribute]` + `HandleCommand` パターン | 同じ |

## 作者

Created by **DevBookOfArray**

[![YouTube](https://img.shields.io/badge/YouTube-DevBookOfArray-red?logo=youtube&logoColor=white)](https://www.youtube.com/@DevBookOfArray)
[![GitHub](https://img.shields.io/badge/GitHub-youngwoocho02-181717?logo=github)](https://github.com/youngwoocho02) (original)
[![GitHub](https://img.shields.io/badge/GitHub-nethunterocean--cmyk-181717?logo=github)](https://github.com/nethunterocean-cmyk/unity-cli) (security fork)

## ライセンス

MIT
