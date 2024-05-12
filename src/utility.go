package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hjson/hjson-go/v4"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
)

// measureExecutionTime は与えられた関数 fn の実行時間を計測します。
// 与えられた関数戻り値を interfase{} で返します。
// 実行例)
// func add(a, b int) int { return a+b }
// ret := MeasureExecutionTime( func() interface{} { return add(1,2) } )
// fmt.Printf("ret: %v\n", ret)
func MeasureExecutionTime(fn func() interface{}) interface{} {
	start := time.Now()
	result := fn()
	elapsed := time.Since(start)
	fmt.Printf("Execution Time: %s\n", elapsed)
	return result
}

func GetText_(filepath string) string {
	ret := GetText(&filepath)
	return *ret
}

func GetText(filepath *string) *string {
	defer func() {
		if err := recover(); err != nil {
			panic(errors.Errorf("%v", err))
		}
	}()

	b, err := os.ReadFile(*filepath) // https://pkg.go.dev/os@go1.20.5#ReadFile
	if err != nil {
		panic(errors.Errorf("Error: %v, file: %v", err, *filepath))
	}
	str := string(b)
	return &str
}

func CreateDirectry(filepath string) {
	if IsExist(filepath) {
		if !IsDir(filepath) {
			//
			panic(errors.Errorf("作成しようとしたディレクトリと同じ名前のファイルが既に存在した。"))
		}
	}
	if err := os.Mkdir(filepath, 0755); err != nil {
		//panic(errors.Errorf("%v", err))
		// エラー無視する。
	}
}

// 指定されたファイルパスの親ディレクトリがなければ作成します。
func CreateParentDir(filePath string) error {
	parentDir := filepath.Dir(filePath)
	// 親ディレクトリが既に存在するかどうかを確認
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		// 親ディレクトリを再帰的に作成
		return os.MkdirAll(parentDir, 0755)
	}
	return nil
}

func GetCurrentDir() string {
	ret, err := os.Getwd()
	if err != nil {
		panic(errors.Errorf("%v", err))
	}
	return filepath.ToSlash(ret)
}

func ChangeDir(dir string) {
	os.Chdir(dir)
}

var gDirStack []string = []string{}

func PushDir(dir string) {
	gDirStack = append(gDirStack, dir)
	os.Chdir(dir)
}
func PopDir() {
	last := len(gDirStack) - 1
	dir := gDirStack[last]
	gDirStack = gDirStack[:last]
	os.Chdir(dir)
}

func AbsPath(path string) string {
	absoluteOutputPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("failed to resolve absolute path for output: %w", err)
		panic(errors.Errorf("failed to resolve absolute path for output: %w", err))
	}
	absoluteOutputPath = filepath.ToSlash(absoluteOutputPath)
	return absoluteOutputPath
}

func WriteText(file, str *string) {

	if dir := filepath.Dir(*file); !IsExist(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(errors.Errorf("Failed to create directory: %w", err))
			return
		}
	}

	f, err := os.Create(*file)
	defer f.Close()
	if err != nil {
		panic(errors.Errorf("%v", err))
	} else {
		if _, err := f.Write([]byte(*str)); err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}

// insertStringAfter inserts the string `insert` after the first occurrence of `str` in `in`.
func insertStringAfter(in, str, insert string) string {
	index := strings.Index(in, str)
	if index == -1 {
		return in // `str` not found in `in`
	}
	// Calculate position to insert the string
	insertionPoint := index + len(str)
	// Create new string with `insert` added
	return in[:insertionPoint] + insert + in[insertionPoint:]
}

func WriteTextBytes(file *string, bs []byte) {
	f, err := os.Create(*file)
	defer f.Close()
	if err != nil {
		panic(errors.Errorf("%v", err))
	} else {
		if _, err := f.Write(bs); err != nil {
			panic(errors.Errorf("%v", err))
		}
	}
}

func CreateTmpDir() string {
	// OSの一時ディレクトリを使用していたが、カレントディレクトリが別ドライブだとその後の処理が大変難しくなるので、
	// カレントディレクトリ直下に作成することにする。
	if false {
		tempDir, err := ioutil.TempDir("", "extracted")
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		return AbsPath(string(tempDir))
	} else {
		t := AbsPath(filepath.Join(GetCurrentDir(), ".tmp"))
		if !IsExist(t) {
			os.RemoveAll(t)
		}
		CreateDirectry(t)
		return t
	}
}

// ExtractEmbeddedFiles はembed.FS型のfilesを一時ディレクトリに展開し、
// 一時ディレクトリのパスを返します。
func ExtractEmbeddedFiles(files embed.FS) (string, error) {
	t := CreateTmpDir()

	err := fs.WalkDir(files, "files", func(path string, d fs.DirEntry, err error) error {
		//RunCommand("pwd")
		//log.Printf("t   : %v", t)
		//log.Printf("path: %v", path)
		//log.Printf("d   : %v", d)
		//log.Printf("err : %v", err)

		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		if d.IsDir() {
			return nil
		}
		data, err := files.ReadFile(path)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}

		//relativePath, err := filepath.Rel(t, path)
		//if err != nil {
		//	panic(errors.Errorf("%v", err))
		//	return err
		//}
		//targetPath := filepath.Join(t, relativePath)
		targetPath := filepath.Join(t, path)

		err = os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		err = ioutil.WriteFile(targetPath, data, d.Type().Perm())
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		return nil
	})
	if err != nil {
		os.RemoveAll(t)
		panic(errors.Errorf("%v", err))
		return "", err
	}
	return t, nil
}

func UnmarshalHjson(b []byte) Setting {
	var setting Setting
	if err := hjson.Unmarshal(b, &setting); err != nil {
		panic(errors.Errorf("%v", err))
	}
	return setting
}

func ReadSettingHjson(path *string) *Setting {

	hjsonStr := GetText(path)
	b := []byte(*hjsonStr)
	//*hjsonStr = fmt.Sprintf("{%v}", *hjsonStr)
	setting := UnmarshalHjson(b)
	//log.Printf("%v", setting)
	var data map[string]interface{}
	if err := hjson.Unmarshal(b, &data); err != nil {
		log.Printf("%v", err)
		panic(errors.Errorf("%v", err))
		//} else {
		/*
			for k, v := range data {
				log.Printf("%v:%v", k, v)
			}
		*/
	}
	//j, _ := json.Marshal([]Setting{setting})
	//log.Printf("setting:\n%v", JsonFormat(j))
	return &setting
}

func Copy_filesindir(old_file, new_file string, embedFileList *EmbedFileList) {
	old_file = filepath.ToSlash(old_file)
	new_file = filepath.ToSlash(new_file)
	if IsDir(old_file) {
		old_file = old_file + "/*"
	}
	Copy(old_file, new_file, embedFileList)
}

func Copy(old_file, new_file string, embedFileList *EmbedFileList) {

	if true {
		old_file = GetRelativePath(old_file)
		new_file = GetRelativePath(new_file)

		cmd := exec.Command(embedFileList.BinBusybox, "cp", "-rf", old_file, new_file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {

		old_file = GetRelativePath(old_file)
		new_file = GetRelativePath(new_file)
		if !IsDir(old_file) {
			if !IsExist(new_file) {
				//CreateDirectry(new_file)
				//new_file = new_file + "/"
			} else if !IsDir(new_file) {
				//new_file = new_file + "/"
			}
		} else {
		}
		runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "cp", "-rf", old_file, new_file)
		return

		cmd := exec.Command("cp", "-rf", old_file, new_file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		return

		oldFile, err := os.Open(old_file)
		if err != nil {
			fmt.Println(err)
			panic(errors.Errorf("%v", err))
		}
		defer oldFile.Close()

		if IsDir(old_file) {
			// oldがディレクトリの時
			if IsExist(new_file) {
				if IsDir(new_file) {
					// 上書きする
				} else {
					// ファイルがあった
					// エラー
					panic(errors.Errorf("ディレクトリをコピー使用と試みましたが、既に同じ名前のファイルが存在しました。: %v", new_file))
				}
			} else {
				CreateDirectry(new_file)
				new_file = filepath.Join(new_file, filepath.Base(new_file))
			}
		} else {
			// oldがディレクトリの時の時
			if IsExist(new_file) {
				if IsDir(new_file) {
					// ディレクトリがあった
					// エラー
					panic(errors.Errorf("ディレクトリをコピー使用と試みましたが、既に同じ名前のファイルが存在しました。: %v", new_file))
				} else {
					// 上書きする
				}
			} else {
				//CreateDirectry(new_file)
			}
		}

		newFile, err := os.Create(new_file)
		if err != nil {
			fmt.Println(err)
			panic(errors.Errorf("%v", err))
		}
		defer newFile.Close()
		io.Copy(newFile, oldFile)
		log.Printf("cp %#v %#v", old_file, new_file)
	}
}

func StringJoin(in []string) *string {
	var m2 = bytes.NewBuffer(make([]byte, 0, 100))
	for _, v := range in {
		m2.WriteString(v)
	}
	str := m2.String()
	return &str
}

func IsDir(path string) bool {
	fInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fInfo.IsDir()
}

func IsExist(path string) bool {
	//info, err := os.Stat(path)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		return false
		//panic(errors.Errorf("%v", err))
	}
	return !os.IsNotExist(err)
}

func getFileNameWithoutExt(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), filepath.Base(path[:len(path)-len(filepath.Ext(path))])))
}

// 拡張子を変える。拡張子の指定は".json"のようにする。
func ChangeFilePathExt(path string, ext string) string {
	return getFileNameWithoutExt(path) + ext
}

// 拡張子を大文字小文字無視して比較する。
func checkExt(path string, exts []string) bool {
	ext := filepath.Ext(path)
	for _, e := range exts {
		if len(e) != 0 && e[0] != '.' {
			e = "." + e // 先頭に.がついてないとき付ける。
		}
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func printOutputWithHeader(header, color string, r io.Reader, verbose bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if verbose {
			fmt.Printf("%s%s\n", header, ansi.Color(scanner.Text(), color))
		}
	}
}

const (
	stdoutColor = "green"
	//stderrColor = "red"
	stderrColor = "magenta"
)

func runCommandOutputRealtimeWithTee(cmd *exec.Cmd, verbose bool, outputFile string) (stdout, stderr string, exitCode int, err error) {
	var file *os.File
	if outputFile != "" {
		// 出力ファイルを作成します
		file, err = os.Create(outputFile)
		if err != nil {
			return "", "", 0, err
		}
		defer file.Close()
	}

	// コマンドの出力先を設定します
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", 0, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return "", "", 0, err
	}

	var bufout, buferr bytes.Buffer
	var outWriter, errWriter io.Writer
	if file != nil {
		outWriter = io.MultiWriter(&bufout, file)
		errWriter = io.MultiWriter(&buferr, file)
	} else {
		outWriter = &bufout
		errWriter = &buferr
	}
	outReader2 := io.TeeReader(outReader, outWriter)
	errReader2 := io.TeeReader(errReader, errWriter)

	if err = cmd.Start(); err != nil {
		return "", "", 0, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() { printOutputWithHeader("stdout:", stdoutColor, outReader2, verbose); wg.Done() }()
	go func() { printOutputWithHeader("stderr:", stderrColor, errReader2, verbose); wg.Done() }()

	wg.Wait()

	err = cmd.Wait()
	stdout = bufout.String()
	stderr = buferr.String()

	if err != nil {
		if err2, ok := err.(*exec.ExitError); ok {
			if s, ok := err2.Sys().(syscall.WaitStatus); ok {
				err = nil
				exitCode = s.ExitStatus()
			}
		}
	}

	return stdout, stderr, exitCode, err
}

//// runCommandOutputTextWithoutPanicByArray はコマンドを実行し、その結果を文字列で返し、
//// 出力を指定されたパスのファイルに保存します。
//func runCommandOutputTextWithoutPanicByArray(output_path string, command string, args ...string) (string, error) {
//	// コマンドを実行して結果を取得
//	stdout, err := exec.Command(command, args...).CombinedOutput()
//	output := string(stdout)
//	if err != nil {
//		return output, err
//	}
//
//	// 出力をファイルに書き込む
//	err = ioutil.WriteFile(output_path, stdout, 0644)
//	if err != nil {
//		return output, err
//	}
//
//	return output, nil
//}

// func runCommandGetOutputWithoutPanicByArray(command string, args ...string) (string, error) {
// 	stdout, err := exec.Command(command, args...).CombinedOutput()
// 	return string(stdout), err
// }

func runCommandByArrayWithoutPanic(command string, args ...string) error {
	//fmt.Printf("%v %v", command, args)
	cmd := exec.Command(command, args...)
	cmd.Stdout = wrapperStdout
	cmd.Stderr = wrapperStderr
	return cmd.Run()
}

// // 1.外部コマンドを実行し、標準出力と標準エラーを印字し、エラーの場合はpanicする
// // この関数では bash "./a b c.sh"のように空白のあるパスに対応できない。
// func RunCommand(command string) {
// 	//panic("")
// 	//log.Printf("run command: %#v\n", command)
// 	tokens := strings.Split(command, " ")
// 	c := tokens[0]
// 	args := tokens[1:]
// 	err := runCommandByArrayWithoutPanic(c, args...)
// 	if err != nil {
// 		log.Printf("run command: %#v\n", command)
// 		panic(errors.Errorf("%v", err))
// 	}
// }

func GetRelativePath(targetPath string) string {
	// カレントディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	// カレントディレクトリから目的のパスまでの相対パスを計算
	relativePath, err := filepath.Rel(currentDir, AbsPath(targetPath))
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	return filepath.ToSlash(relativePath)
}

// ListEmbeddedFiles は埋め込まれたファイルのリストを出力します。
func ListEmbeddedFiles(dir string) {
	// 'data' ディレクトリ内のファイルをリストアップ

	dirEntries, err := files.ReadDir(dir)
	if err != nil {
		fmt.Println("ファイルのリストアップ中にエラーが発生しました:", err)
		return
	}

	// ファイル名を出力
	for _, entry := range dirEntries {
		fmt.Println(entry.Name())
	}
}

func UnZip(zipFile, destDir string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	destDir = filepath.Clean(destDir)

	extractFile := func(file *zip.File, destDir string) error {
		destPath := filepath.Join(destDir, file.Name)
		destPath = filepath.Clean(destPath)

		// Check for file traversal attack
		if !strings.HasPrefix(destPath, destDir) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				panic(errors.Errorf("%v", err))
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				panic(errors.Errorf("%v", err))
			}

			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				panic(errors.Errorf("%v", err))
			}
			defer destFile.Close()

			srcFile, err := file.Open()
			if err != nil {
				panic(errors.Errorf("%v", err))
			}
			defer srcFile.Close()

			if _, err := io.Copy(destFile, srcFile); err != nil {
				return err
			}
		}

		return nil
	}

	for _, file := range reader.File {
		if err := extractFile(file, destDir); err != nil {
			panic(errors.Errorf("%v", err))
		}
	}

	return nil
}

// ConvertJSONToHJSON takes a JSON string and returns its HJSON representation.
func ConvertJSONToHJSON(jsonStr string) (string, error) {
	var dat map[string]interface{}

	// Parse JSON string into a map.
	if err := hjson.Unmarshal([]byte(jsonStr), &dat); err != nil {
		return "", err
	}

	// Convert map back to HJSON string.
	hjsonStr, err := hjson.Marshal(dat)
	if err != nil {
		return "", err
	}

	return string(hjsonStr), nil
}

// ConvertJSONToHJSON takes a JSON string and returns its HJSON representation.
func ConvertJSONdataToHJSON(jsonData []byte) (string, error) {
	var dat map[string]interface{}

	// Parse JSON string into a map.
	if err := hjson.Unmarshal(jsonData, &dat); err != nil {
		return "", err
	}

	// Convert map back to HJSON string.
	hjsonStr, err := hjson.Marshal(dat)
	if err != nil {
		return "", err
	}

	return string(hjsonStr), nil
}

// RemoveEmptyLines removes all empty lines from the given string.
func RemoveEmptyLines(input string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return strings.TrimRight(result.String(), "\n") // Remove the last newline added
}

// addTabToLines takes a string, splits it by new lines, and prepends a tab to each line.
func addTabToLines(input string) string {
	lines := strings.Split(input, "\n") // Split the input into lines.
	var result strings.Builder          // Use a Builder to efficiently build the new string.
	for _, line := range lines {
		result.WriteString("\t" + line + "\n") // Append a tab to the start of each line.
	}
	return result.String()
}

// chatgpt製 最も外側の{}の内側をインデントする関数。
// IndentOutermostBraces takes a string, splits it into lines, and indents the contents between the outermost braces.
func IndentOutermostBraces(input string) string {
	lines := strings.Split(input, "\n")
	var result strings.Builder
	inBraces := false // Flag to indicate whether we are currently between the outermost braces.

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Contains(trimmedLine, "{") && !inBraces {
			// Opening brace of the outermost block found.
			inBraces = true
			result.WriteString(line + "\n")
			continue
		}
		if inBraces {
			if strings.Contains(trimmedLine, "}") {
				// Closing brace of the outermost block found.
				inBraces = false
				result.WriteString(line + "\n")
				continue
			}
			// Add a tab to the start of the line since this line is inside the outermost braces.
			result.WriteString("\t" + line + "\n")
		} else {
			// Not inside braces, just copy the line.
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}
