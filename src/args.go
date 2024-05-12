package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexflint/go-arg"
)

type Args struct {
	CreateEmptyHjson *ArgsCreateEmptyHjson `arg:"subcommand:init"     help:"空の設定ファイルを生成する。"`
	ConvertToJson    *ArgsConvert          `arg:"subcommand:convert"  help:"設定ファイルをjsonに変換する。スクリプト内でjsonをjqで参照する、のような用途を想定。"`
	GenScript        *ArgsScript           `arg:"subcommand:generate" help:"設定ファイルに記述したスクリプトからshellscirptを生成する。既にあれば上書きされるので注意。"`
	RunScript        *ArgsRunScript        `arg:"subcommand:run"      help:"生成されたスクリプトを実行する。引数に実行する設定ファイルに記載したスクリプト名を指定する。"`
	RunAll           *ArgsRunAll           `arg:"subcommand:run-all"  help:"適切な設定ファイルがある前提で、get,merge,script-gen,script-run,createを一括で実行する。"`
	OutputSourceCode *ArgsOutputSrc        `arg:"subcommand:code"     help:"このプログラムのソースコードを出力する。golangのコンパイラは https://go.dev/dl/ からダウンロードできる。"`
	VersionSub       *ArgsVersion          `arg:"subcommand:version"  help:"バージョン番号を出力する。-vと同じ。" `

	CheckSetting       bool   `arg:"--check-setting"                     help:"設定ファイルを読みこんでどのように解釈されたかを出力する。"`
	SettingPath        string `arg:"-s,--setting"                        help:"使用したい設定ファイルパス指定。指定がなければ./setting.hjsonを使用する。" placeholder:"FILE"`
	CurrentDir         string `arg:"-c,--cd"                             help:"カレントディレクトリを指定のディレクトリに変更してから実行する。"`
	Readme             bool   `arg:"-r,--show-readme"                    help:"readme.mdを出力する。"`
	LogPath            string `arg:"-l,--logfile"                        help:"ログファイル出力先を指定する。設定ファイルで指定されておらず、この指定がないときログ出力しない。" placeholder:"FILE"`
	Silent             bool   `arg:"-q,--quiet"                          help:"標準出力と標準エラー出力に出力しない。"`
	DebugRetrainTmpDir bool   `arg:"-t,--retrain-tmp"                    help:"実行時に生成した一時ディレクトリを削除しない。\n                         script-genなどで生成したshellscriptを編集した後に、一時ファイル内のbusyboxやmkisofsなどを利用できる。"`
	Version            bool   `arg:"-v,--version"                        help:"バージョン番号を出力する。サブコマンドversionと同じ。" `
	Help               bool

	//BashScript         *ArgsBashScript  `arg:"subcommand:bash"         help:"busybox bash を呼び出す。"`
	//YaegiScript        *ArgsYaegiScript `arg:"subcommand:yaegi"        help:"yaegi を呼び出す。"`

}

func (Args) Description() string {
	return fmt.Sprintf(`%v version %v.%v

Description:
  複数のshellscriptを修正して実行といった作業の自動化補助ツール。
  この設定ファイルを使用して、複数個のshellscriptを、設定ファイル内の値を用いて自動生成して、一括または個別実行できる。
`,
		filepath.Base(os.Args[0]), version, revision)

}

// ShowHelp() で使う
var parser *arg.Parser

func ShowHelp() {
	buf := new(bytes.Buffer)
	parser.WriteHelp(buf)
	fmt.Printf("%v\n", buf.String())
	os.Exit(1)
}

// 引数解析
func ArgParse() (*Args, *Setting) {
	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する

	args := &Args{
		SettingPath: "./setting.hjson",
	}

	var err error
	//parser = arg.MustParse(args)

	{
		n := filepath.Base(filepath.ToSlash(os.Args[0]))
		//log.Printf("name: %v", n)
		parser, err = arg.NewParser(arg.Config{Program: n}, args)
		if err != nil {
			log.Printf("%v", err)
		}
		if len(os.Args) == 1 {
			args.Help = true
			args.CurrentDir = "" // ShowHelpが呼ばれるとき空にしておく。
			ShowHelp()
		} else if err = parser.Parse(os.Args[1:]); err != nil {
			//log.Printf("%v", err)
		}
		// --versionがなぜかtrueにならないので仕方なくチェック
		for _, arg := range os.Args[1:] {
			if arg == "--version" {
				args.Version = true
				break
			}
			if arg == "--help" || arg == "-h" {
				args.Help = true
				ShowHelp()
			}
		}
		//args.Print()
	}

	if args.Version || args.VersionSub != nil {
		fmt.Printf("%v version %v.%v\n", filepath.Base(os.Args[0]), version, revision)
		os.Exit(1)
	}

	var s *Setting
	if args.OutputSourceCode != nil || //
		false {
		// サブコマンドで設定ファイルが必要出ないコマンドの時設定ファイルを読みこまない
	} else {
		s, err = ReadSetting(args, s)
		loggingSettings(args, s)
		if args.CheckSetting {
			log.Printf(s.Print(args))
		}
	}

	if args.CheckSetting {
		args.Print()
	}

	// defer用にグローバル変数に設定する。
	if args.DebugRetrainTmpDir {
		gDebugRetrainTmpDir = true
	}

	// 引数でカレントディレクトリを指定された場合移動。
	if len(args.CurrentDir) != 0 {
		ChangeDir(args.CurrentDir)
		CurrentPath = args.CurrentDir
	}

	if args.CreateEmptyHjson != nil {
		CreateEmptyHjson(args)
		os.Exit(0)
	}

	if args.Silent {
		// 引数で黙るように指定されたとき出力先をnullにする。
		//stdout := os.Stdout
		//defer func() { os.Stdout = stdout }()
		os.Stdout = os.NewFile(0, os.DevNull)
		os.Stderr = os.NewFile(0, os.DevNull)
	}

	// 引数として必須な何れかが欠けている場合ヘルプを出力する。

	if args.Version == false && args.VersionSub == nil && //
		args.CheckSetting == false && //
		args.GenScript == nil && args.RunScript == nil && args.RunAll == nil && //
		args.OutputSourceCode == nil && //
		args.ConvertToJson == nil && //
		args.DebugRetrainTmpDir == false && args.Readme == false && //
		//args.BashScript == nil && //
		//args.YaegiScript == nil && //
		true {
		args.Help = true
	}

	if args.Help {
		ShowHelp() // go-argsの生成するヘルプ文字列を取得して出力する。
	}

	return args, s
}

// ヘルパー関数で関数名と行番号を取得し、フォーマットして返す
func LogCallerInfo() {
	pc, file, line, ok := runtime.Caller(1) // Caller(1) は呼び出し元の情報を取得
	if !ok {
		//return "情報を取得できませんでした"
		return
	}
	funcName := runtime.FuncForPC(pc).Name() // PC値から関数名を取得
	funcName = path.Base(funcName)           // フルパスからベース名のみ取得
	fileName := path.Base(file)              // フルパスからファイル名のみ取得
	//str := fmt.Sprintf("%s:%d %s", fileName, line, funcName)
	//fmt.Printf("%s %s:%d %s\n", time.Now().Format("15:04:05"), fileName, line, funcName)
	fmt.Printf("%s %s:%d \n", time.Now().Format("15:04:05"), fileName, line)

}

func (args *Args) Print() { // {{{
	log.Printf(`
	GenScript             : %#v
	RunScript             : %#v
	RunAll                : %#v
	OutputSourceCode      : %#v
	CheckSetting          : %#v
	SettingPath           : %#v
	CurrentDir            : %#v
	Readme                : %#v
	LogPath               : %#v
	Silent                : %#v
	DebugRetrainTmpDir    : %#v
	Version               : %#v
	VersionSub            : %#v
	Help                  : %#v
	`,
		args.GenScript,
		args.RunScript,
		args.RunAll,
		args.OutputSourceCode,
		args.CheckSetting,
		args.SettingPath,
		args.CurrentDir,
		args.Readme,
		args.LogPath,
		args.Silent,
		args.DebugRetrainTmpDir,
		args.Version,
		args.VersionSub,
		args.Help,
	)
} // }}}

type ArgsRunAll struct {
}
type ArgsCreateTemplateSetting struct {
}
type ArgsMerge struct {
	//DstDirNamePostfix []string `arg:"dst_dir_name_postfix"`
}
type ArgsScript struct{}
type ArgsRunScript struct {
	ScriptName []string `arg:"positional"`
}
type ArgsOutputSrc struct{}
type ArgsVersion struct{}
type ArgsConvert struct {
	//Tab    bool   `arg:"--tab"   help:"jsonを見やすくformatして出力する。インデントにtabを使用する。"`
	Mini   bool   `arg:"-m,--mini" help:"minifyしたjsonを出力する。"`
	Space  int    `arg:"--space"   help:"jsonを見やすくformatして出力する。インデントに指定個数の半角空白を使用する。負の数の時無視される。" default:"-1"`
	Output string `arg:"--output"  help:"指定のパスにテキストファイルとして出力する。"`
	Color  bool   `arg:"--color" help:"色を付ける。そのあとにパイプで処理するとバグる。"`
}
type ArgsCreateEmptyHjson struct{}
