package main

import (
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
)

var wrapperStdout io.Writer
var wrapperStderr io.Writer
var logfile *os.File

func loggingSettings(args *Args, s *Setting) {
	// 設定ファイルが存在、かつ設定ファイル中でフラグが経っていたら立てる。
	// 設定ファイルで指定されていなくても、引数で指定されていたら立てる。設定ファイル指定よりも引数指定のほうが強い
	flag_silent := (s != nil && s.Silent) || args.Silent

	logpath := func() string {
		if len(args.LogPath) != 0 {
			return args.LogPath
		} else if s != nil && len(s.LogPath) != 0 {
			return s.LogPath
		}
		return ""
		//return filepath.ToSlash(filepath.Join(GetCurrentDir(), fmt.Sprintf("%v.log", getFileNameWithoutExt(filepath.Base(os.Args[0])))))
	}()
	if len(logpath) != 0 {
		//log.Printf("log filepath : %v", logpath)
		var err error
		logfile, err = os.OpenFile(logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}

		if flag_silent {
			wrapperStdout, wrapperStderr = logfile, logfile // 出力をログファイルにのみ出力する。
		} else {
			wrapperStdout, wrapperStderr = io.MultiWriter(os.Stdout, logfile), io.MultiWriter(os.Stderr, logfile) // 出力をログファイル、標準出力、標準エラー出力に出力する。
		}
	} else {
		if flag_silent {
			wrapperStdout, wrapperStderr = io.Discard, io.Discard // 出力を全て破棄する
		} else {
			wrapperStdout, wrapperStderr = os.Stdout, os.Stderr // 出力を標準出力、標準エラー出力に出力する。
		}
	}
	log.SetOutput(wrapperStdout)             // logの出力先
	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する
}
