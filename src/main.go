package main

// bincodeのように、バイナリだけになってもソースコードが得られるようにすべてのファイルを埋め込む。
// 一時ディレクトリを作成してそこにこのリポジトリのファイルを丸っとコピー。
// その後/src/embedにリネームする。
// 何でこんなことをするかというと、go:embedはパッケージの親ディレクトリのファイルを埋め込めないため。

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	version         string = "debug"
	revision        string
	ThisProgramPath string // $0の絶対パスを入れる
	CurrentPath     string // 起動時のカレントディレクトリ

	developing bool = false
)

func main() {
	CurrentPath = GetCurrentDir()
	ThisProgramPath = AbsPath(os.Args[0])

	args, s := ArgParse()

	if developing {
		// 開発用の一時的なコード
		var s *Setting
		var err error
		s, err = ReadSetting(args, s)
		if err != nil {
			log.Printf("%v", err)
		}
		embedFileList := LoadEmbedFiles() // 埋め込んだファイル群を一時ディレクトリにロード 展開にそれ相応の時間がかかる。
		defer DeleteTmpDir(embedFileList) // 終了時に一時ディレクトリを削除する。
		//log.Println(GenScriptHead(args, s, embedFileList))
		return
	}
	Run(args, s)
}

func Run(args *Args, s *Setting) {

	embedFileList := LoadEmbedFiles() // 埋め込んだファイル群を一時ディレクトリにロード 展開にそれ相応の時間がかかる。
	defer DeleteTmpDir(embedFileList) // 終了時に一時ディレクトリを削除する。

	switch {
	case args.ConvertToJson != nil:
		GenJson(args, s, embedFileList) // jsonを生成

	case args.OutputSourceCode != nil:
		//このプログラムのソースコードを出力する。
		runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "rm", "-rf", "./create_installer_iso_src")
		runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "mkdir", "-p", "./create_installer_iso_src")
		src := GetRelativePath(filepath.Join(embedFileList.TmpDir, "create_installer_iso.tar"))
		runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "tar", "xf", src, "-C", "./create_installer_iso_src")
		fmt.Printf("%q \nに出力しました。\n", AbsPath("./create_installer_iso_src"))

	case args.GenScript != nil:
		if s == nil {
			log.Printf("設定ファイルがありません")
		}
		fmt.Printf("run GenScript\n")
		GenScript(args, s, embedFileList)

	case args.RunScript != nil:
		if s == nil {
			log.Printf("設定ファイルがありません")
		}
		fmt.Printf("run RunScript\n")
		RunScript(args, s, embedFileList)

	case args.RunAll != nil:
		fmt.Printf("run RunAll\n")
		RunAll(args, s, embedFileList)

	case args.Readme: // readme.mdを出力
		readme := GetText(&embedFileList.Readme)
		fmt.Printf("%s", *readme)

	}
	return
}
