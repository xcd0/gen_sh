package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

func GenScriptHead(args *Args, s *Setting, embedFileList *EmbedFileList) string { // {{{
	script := ""
	script = fmt.Sprintf(`%v
####################################################################################################
# 設定ファイル内のshellscriptではjqがそのまま使える。
#   $ ${args_this_program_path} convert --output "./setting.json"
#   $ jq ".log" "./setting.json"
# のようにすると設定ファイル内のキーがlogの文字列が得られる。
# jqの詳しい使い方は https://www.tohoho-web.com/ex/jq.html などを参照。
args_this_program_path=%#v
args_setting_hjson=%#v
args_log=%#v
args_cd=%#v
setting_log=%#v
`, script, //
		ThisProgramPath, args.SettingPath, args.LogPath, args.CurrentDir, s.LogPath)
	//if len(s.Constant) != 0 {
	//	script = fmt.Sprintf("%v\n# 設定ファイルで定義された定数 ここから", script)
	//	for k, v := range s.Constant {
	//		script = fmt.Sprintf("%v\n%v=\"%v\"", script, k, v)
	//	}
	//	script = fmt.Sprintf("%v\n# 設定ファイルで定義された定数 ここまで\n", script)
	//	//log.Printf("%v", script)
	//}
	// cd "$(dirname "$0")"
	// はダメ。というのもscriptが生成される場所は別の場所なため。
	script = fmt.Sprintf(`%v
pwd
current_dir=$(pwd)
scrit_dir="$(cd "$(dirname "$0")"; pwd)"

# 実行ファイルに埋め込まれているバイナリのあるディレクトリにパスを通す。
bin_dir="%v"
export PATH="${bin_dir}:$PATH"
####################################################################################################
`, script,
		filepath.ToSlash(filepath.Join(embedFileList.TmpDir, "files/bin")), // ここにPATHを通す。
	)
	script = ReplaceCommand(script, embedFileList)
	return script
} // }}}

func GenScriptFromString(script string, output_name string, args *Args, s *Setting, embedFileList *EmbedFileList) {
	log.Printf("GenScriptFromString: %v\n---\n%v\n---", output_name, script)
	if len(script) > 0 {
		header := GenScriptHead(args, s, embedFileList)
		script = ReplaceCommand(script, embedFileList) // 設定ファイルのスクリプトに含まれるコマンドのうち、バイナリに埋め込んでいるコマンドの置き換え。
		script = addTabToLines(script)                 // 設定ファイルのスクリプトを1つインデントする。

		output := fmt.Sprintf(`#!/bin/sh
%v
# %v
script_name=%v
{
	%v
} | sed 's/^/'$script_name'> /'
`, header, output_name, output_name, script)

		p := AbsPath(filepath.Join(s.DstTmp, output_name))
		os.RemoveAll(p)
		WriteText(&p, &output)
	}
}

func GenScript(args *Args, s *Setting, embedFileList *EmbedFileList) {
	//log.Printf("GenScript")
	//log.Printf("s.ScriptForDst: %v", s.ScriptForDst)

	for _, k := range s.Scripts.Keys() {
		v, _ := s.Scripts.Get(k)
		GenScriptFromString(v.(string), k, args, s, embedFileList)
	}
	return
}

func RunScript(args *Args, s *Setting, embedFileList *EmbedFileList) {

	pwd := AbsPath(s.DstTmp)
	log.Printf("RunScript: ChangeDir(%#v)", pwd)

	log.Printf("args.RunScript.ScriptName : %v", args.RunScript.ScriptName)

	for _, k := range args.RunScript.ScriptName {
		ChangeDir(pwd)
		v, _ := s.Scripts.Get(k)
		log.Printf("RunScript: %v : \n---\n%v\n---", k, v)
		p := AbsPath(k)
		t := ChangeFilePathExt(p, ".log")
		log.Printf("RunScript: run: bash %#v", p)
		log.Printf("t       : %#v", t)
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "time", "bash", "-x", p), true, t)
		log.Printf("%v", GetText_(t))
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}

/*
func RunScript(s *Setting, embedFileList *EmbedFileList) {
	RunScriptForDst(s, embedFileList)
	RunScriptForSpecificDisk(s, embedFileList)
}

func RunScriptWithNameOnTmp(name string, args *Args, s *Setting, embedFileList *EmbedFileList) {
	p := AbsPath(filepath.Join(s.DstTmp, name))
	if !IsExist(p) {
		return
	}
	log.Printf("RunScript: ChangeDir(%#v)", s.DstTmp)
	ChangeDir(s.DstTmp)
	log.Printf("RunScript: %v: \n---\n%v\n---", p, s.ScriptForDst)
	t := ChangeFilePathExt(p, ".log")
	_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "time", "bash", "-x", p), true, t)
	log.Printf("%v", GetText_(t))
	if err != nil {
		panic(errors.Errorf("%v", err))
	}
}

func RunScriptForDst(s *Setting, embedFileList *EmbedFileList) {
	pwd := AbsPath(s.DstTmp)
	log.Printf("RunScript: ChangeDir(%#v)", pwd)
	ChangeDir(pwd)
	if len(s.ScriptForDst) != 0 {
		log.Printf("RunScript: s.ScriptForDst : \n---\n%v\n---", s.ScriptForDst)
		p := AbsPath("script_for_dst.sh")
		t := ChangeFilePathExt(p, ".log")
		log.Printf("s.DstTmp: %#v", s.DstTmp)
		log.Printf("RunScript: run: bash %#v", p)
		log.Printf("t       : %#v", t)
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "time", "bash", "-x", p), true, t)
		log.Printf("%v", GetText_(t))
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}

func RunScriptForSpecificDisk(s *Setting, embedFileList *EmbedFileList) {
	pwd := AbsPath(s.DstTmp)
	log.Printf("RunScript: ChangeDir(%#v)", pwd)
	ChangeDir(pwd)
	for i := 0; i < len(s.ScriptForSpecificDisk); i++ {
		log.Printf("RunScript: s.ScriptForSpecificDisk[%v] : %#v, len: %v", i, s.ScriptForSpecificDisk[i], len(s.ScriptForSpecificDisk[i]))
		if len(s.ScriptForSpecificDisk[i]) == 0 {
			continue
		}
		//p := AbsPath(filepath.Join(pwd, fmt.Sprintf("script_for_specific_%v.sh", s.OutputIsoDir[i])))
		//p := AbsPath(filepath.Join(pwd, fmt.Sprintf("script_for_specific_%v.sh", i)))
		p := AbsPath(fmt.Sprintf("script_for_specific_%v.sh", i))
		log.Printf("RunScript: bash %#v", p)
		t := ChangeFilePathExt(p, ".log")
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "time", "bash", "-x", p), true, t)
		log.Printf("%v", GetText_(t))
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}
*/
