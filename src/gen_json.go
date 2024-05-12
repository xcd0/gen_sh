package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hjson/hjson-go/v4"
	"github.com/pkg/errors"
)

func GenJson(args *Args, s *Setting, embedFileList *EmbedFileList) {

	if !fileExists(args.SettingPath) {
		panic(errors.Errorf("設定ファイル %v がありませんでした。\n", args.SettingPath))
	}
	b, err := os.ReadFile(args.SettingPath)
	if err != nil {
		log.Fatalf("設定ファイル %v が見つかりませんでした。`create_installer_iso.exe template`で設定ファイルの雛形を出力できます。", args.SettingPath)
		ShowHelp()
	}

	if false {
		var node hjson.Node
		if err := hjson.Unmarshal(b, &node); err != nil {
			panic(errors.Errorf("%v", err))
		}
		jsonData, err := node.MarshalJSON()
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		p := ChangeFilePathExt(args.SettingPath, ".json")
		WriteTextBytes(&p, jsonData)
	} else {
		// 構造体を整形されたJSON文字列に変換
		var err error
		var jsonData []byte
		if args.ConvertToJson != nil {
			if args.ConvertToJson.Mini {
				jsonData, err = json.Marshal(*s)
				dst := &bytes.Buffer{}
				if err := json.Compact(dst, jsonData); err != nil {
					panic(err)
				}
				jsonData = dst.Bytes()
			} else if args.ConvertToJson.Space >= 0 {
				jsonData, err = json.MarshalIndent(*s, "", strings.Repeat(" ", args.ConvertToJson.Space))
			} else {
				jsonData, err = json.MarshalIndent(*s, "", "\t")
			}
		}
		if err != nil {
			// エラー処理
			fmt.Println("Error marshalling JSON:", err)
			panic(errors.Errorf("%v", err))
			return
		}

		output := args.ConvertToJson.Output
		if len(output) == 0 {
			output = ChangeFilePathExt(args.SettingPath, "json")
			if IsExist(output) {
				output = getFileNameWithoutExt(output) + "_converted.json"
			}
		}
		WriteTextBytes(&output, jsonData)
		//_, _, _, err = runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinJq, ".", output), true, "")
		if true {
			if args.ConvertToJson.Color {
				err = runCommandByArrayWithoutPanic(embedFileList.BinJq, "--tab", "-C", ".", output)
			} else {
				err = runCommandByArrayWithoutPanic(embedFileList.BinJq, "--tab", ".", output)
			}
			err = runCommandByArrayWithoutPanic(embedFileList.BinJq, "--tab", ".", output)
		} else {
			script := fmt.Sprintf(`#!/bin/bash
export PATH="%v:$PATH"
cat %#v | jq --tab -C .
`,
				filepath.ToSlash(filepath.Join(embedFileList.TmpDir, "files/bin")), // ここにPATHを通す。
				output,
			)
			WriteTextBytes(&output, jsonData)
			p := filepath.ToSlash(filepath.Join(s.DstTmp, "script_for_specific_.sh"))
			WriteText(&p, &script)
			err = runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "bash", p)
		}

		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		if len(args.ConvertToJson.Output) == 0 {
			os.RemoveAll(output)
		}
	}

	return
}
