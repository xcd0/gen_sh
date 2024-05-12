package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

func RunAll(args *Args, s *Setting, embedFileList *EmbedFileList) {

	start := time.Now() // 実行開始時刻

	GenScript(args, s, embedFileList)

	pwd := AbsPath(s.DstTmp)
	log.Printf("RunScript: ChangeDir(%#v)", pwd)

	for _, k := range s.Scripts.Keys() {
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

	fmt.Println("finish.")

	fmt.Printf("\nreal : %v\n", formatTime(time.Since(start))) // 経過時間の出力

}

// time.Durationをtimeコマンドの形式にフォーマットする関数
func formatTime(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("\n%dm%.3fs", minutes, seconds)
}
