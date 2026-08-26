# GolangTextEditor

Go言語で2Dゲームエンジンをプログラムベースで実装するための、テキストエディタと
IDEの中間に位置する、軽量・高速なエディタ（開発中）。

- 対応OS: Windows / Linux
- 実装言語: Go のみ
- 描画/入力バックエンド: 現在は [ebiten](https://ebitengine.org/) を採用
  （将来的に go-sdl2 / raylib-go / go-gl への切替・追加を検討）

現在のフェーズは **基本的なテキストエディタ機能** の実装です。
ゲームエンジン統合などは今後のフェーズで扱います。

プロジェクトの設計方針・遵守ルールは以下を参照してください。

- [`CLAUDE.md`](./CLAUDE.md) — アーキテクチャ方針・コーディング規約
- [`.claude/skills/golang-editor-rules/SKILL.md`](./.claude/skills/golang-editor-rules/SKILL.md) — 絶対遵守チェックリスト

## ビルド・実行

```sh
go build -o editor ./cmd/editor
./editor path/to/file.txt   # 引数を省略すると空のバッファで起動
```

Windows向けクロスコンパイル例（Linux上から、CGO不要）:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o editor.exe ./cmd/editor
```

## 操作方法（フェーズ1時点）

- 文字入力・矢印キーでのカーソル移動・Enter・Backspace・Delete・Home/End
- `Ctrl+S`（または `Cmd+S`）で保存（ファイルを指定して起動した場合のみ）

## テスト

```sh
go test ./...
```

`internal/buffer`（テキストバッファ本体）と `internal/editor`（編集コマンド・
カーソル・ファイルI/O）はディスプレイ非依存で、ヘッドレス環境でもテスト可能です。
