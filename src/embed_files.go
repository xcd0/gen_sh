package main

import (
	"archive/tar"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var _debug bool = true

//go:embed embed.tar
var files embed.FS

type EmbedFileList struct {
	TmpDir string
	Tar    string

	//TemplateSetting           string
	//TemplateSetting        string
	TemplateSettingSimple  string
	TemplateSettingComplex string
	TestData               string

	Readme string

	BinMkisofs  string
	BinJq       string
	Bin7zip     string
	BinNkf      string
	BinRealpath string
	//BinDiff    string // busyboxのdiffはしょぼい
	BinDifft string // busyboxのdiffはしょぼい
	// BinTree    string
	BinRsync   string
	BinYaegi   string
	BinBusybox string
}

// var embedFileList *EmbedFileList

// 一時ディレクトリに埋め込まれたtarファイルを配置する関数
func ExtractEmbeddedTar() (string, error) {
	tempDir := CreateTmpDir()
	t := "embed.tar"

	// 埋め込まれたファイルを読み込む
	fileData, err := fs.ReadFile(files, t)
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	// 一時ディレクトリにファイルを書き込む
	tempFilePath := filepath.Join(tempDir, t)
	err = ioutil.WriteFile(tempFilePath, fileData, 0644)
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	// tarを展開する
	if err := untar(tempFilePath, tempDir); err != nil {
		panic(err)
	}

	return tempDir, nil
}

// 実行ファイルに埋め込んだファイル群を一時ディレクトリに展開してそのパスを構造体に保持する。
func LoadEmbedFiles() *EmbedFileList {

	var embedFileList *EmbedFileList
	// tarにまとめて埋め込むようにした。
	tmpdir, err := ExtractEmbeddedTar()
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	//fmt.Println("------------------------------------")

	tmpdir = AbsPath(tmpdir)

	embedFileList = &EmbedFileList{
		TmpDir: tmpdir,
		Tar:    filepath.ToSlash(filepath.Join(tmpdir, "embed.tar")),
		//TemplateSetting: filepath.ToSlash(filepath.Join(tmpdir, "files/template/setting.hjson")),
		TemplateSettingSimple:  filepath.ToSlash(filepath.Join(tmpdir, "files/template/setting_simple_example.hjson")),
		TemplateSettingComplex: filepath.ToSlash(filepath.Join(tmpdir, "files/template/setting_complex_example.hjson")),
		TestData:               filepath.ToSlash(filepath.Join(tmpdir, "files/test data")), // テストデータを配置する。空白を含むパスで問題が出るか確かめるために敢えてディレクトリに空白を含めている。
		Readme:                 filepath.ToSlash(filepath.Join(tmpdir, "readme.md")),
		BinMkisofs:             filepath.ToSlash(filepath.Join(tmpdir, "files/bin/mkisofs.exe")),
		BinJq:                  filepath.ToSlash(filepath.Join(tmpdir, "files/bin/jq.exe")),
		Bin7zip:                filepath.ToSlash(filepath.Join(tmpdir, "files/bin/7z.exe")),
		BinNkf:                 filepath.ToSlash(filepath.Join(tmpdir, "files/bin/nkf.exe")),
		BinRealpath:            filepath.ToSlash(filepath.Join(tmpdir, "files/bin/realpath.exe")),
		//BinDiff:         filepath.ToSlash(filepath.Join(tmpdir, "files/bin/diff.exe")),
		BinDifft: filepath.ToSlash(filepath.Join(tmpdir, "files/bin/difft.exe")),
		//BinTree:         filepath.ToSlash(filepath.Join(tmpdir, "files/bin/tree.exe")),
		BinRsync:   filepath.ToSlash(filepath.Join(tmpdir, "files/bin/rsync.exe")),
		BinYaegi:   filepath.ToSlash(filepath.Join(tmpdir, "files/bin/yaegi.exe")),
		BinBusybox: filepath.ToSlash(filepath.Join(tmpdir, "files/bin/busybox64u.exe")),
	}
	// if false {
	// 	//runCommandByArrayWithoutPanic(embedFileList.BinTree, tmpdir)
	// 	cmd := exec.Command(embedFileList.BinTree, tmpdir)
	// 	cmd.Stdout = wrapperStdout
	// 	cmd.Stderr = wrapperStderr
	// 	if err := cmd.Run(); err != nil {
	// 		panic(errors.Errorf("%v", err))
	// 	}
	// }
	return embedFileList
}

// 雑に置換しすぎなので、例えばスクリプト内のコメントなどに偶々マッチする文字列があったら置換してしまう。
func ReplaceCommand(str string, embedFileList *EmbedFileList) string {
	if !_debug {
		str = strings.ReplaceAll(str, "bash", fmt.Sprintf("%#v bash", embedFileList.BinBusybox))
		str = strings.ReplaceAll(str, "mkisofs", fmt.Sprintf("%#v", embedFileList.BinMkisofs))
		str = strings.ReplaceAll(str, "jq", fmt.Sprintf("%#v", embedFileList.BinJq))
		str = strings.ReplaceAll(str, "7z", fmt.Sprintf("%#v", embedFileList.Bin7zip))
		str = strings.ReplaceAll(str, "nkf", fmt.Sprintf("%#v", embedFileList.BinNkf))
		str = strings.ReplaceAll(str, "realpath", fmt.Sprintf("%#v", embedFileList.BinRealpath))
		//str = strings.ReplaceAll(str, "diff", fmt.Sprintf("%#v", embedFileList.BinDiff))
		str = strings.ReplaceAll(str, "difft", fmt.Sprintf("%#v", embedFileList.BinDifft))
		//str = strings.ReplaceAll(str, "tree", fmt.Sprintf("%#v", embedFileList.BinTree))
		str = strings.ReplaceAll(str, "rsync", fmt.Sprintf("%#v", embedFileList.BinRsync))
		str = strings.ReplaceAll(str, "yaegi", fmt.Sprintf("%#v", embedFileList.BinYaegi))
		return str
	} else {
		return str
	}
}

var gDebugRetrainTmpDir bool = false

func DeleteTmpDir(embedFileList *EmbedFileList) {

	if false {
		log.Printf("gDebugRetrainTmpDir : %v", gDebugRetrainTmpDir)
	}

	if !gDebugRetrainTmpDir {
		if logfile != nil {
			logfile.Close()
		}
		// 終了時に一時ディレクトリを削除する。
		if embedFileList != nil {
			if false {
				runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "ls", embedFileList.TmpDir)
			}
			err := os.RemoveAll(embedFileList.TmpDir)
			if err != nil {
				// 500ms後に1度だけリトライ
				time.Sleep(500 * time.Millisecond)
				err := os.RemoveAll(embedFileList.TmpDir)
				if err != nil {
					time.Sleep(500 * time.Millisecond)
					panic(errors.Errorf("%v", err))
				}
			}
		}
	}
}

// tarファイルを展開する関数
func untar(tarFile, targetDir string) error {
	// tarファイルを開く
	file, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// tarリーダを作成
	tarReader := tar.NewReader(file)

	// tarファイル内の各エントリに対してループ
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // ファイルの終わりに達した
		}
		if err != nil {
			return err
		}

		// ファイルのフルパスを取得
		path := filepath.Join(targetDir, header.Name)

		// ファイルタイプに応じて処理
		switch header.Typeflag {
		case tar.TypeDir:
			// ディレクトリの場合、ディレクトリを作成
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// 通常ファイルの場合、ファイルを作成
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}

			// ファイルに内容をコピー
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
