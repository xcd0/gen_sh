package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/hjson/hjson-go/v4"
	"github.com/pkg/errors"
)

// Scripts   map[string]string     `json:"scripts"                  comment:"生成するスクリプトを記述する。\nbusyboxのbashで実行される。\n\"スクリプト名\":\"スクリプト実装\" のように記述する。"`

type Setting struct {
	//Constant     map[string]string `json:"constant"                 comment:"設定ファイル内の全てのスクリプト内で使用する定数を定義する。\nスクリプト内で共通の定数などを定義したいときに定義する。\n定数定義の中で定数や変数は使用できない。"`
	ScriptHeader string        `json:"script-header"            comment:"生成するスクリプト全ての先頭に付加するheader。\n変数定義などに使用できる。constantの後に挿入される。"`
	Scripts      *OrderedMapSS `json:"scripts"                  comment:"生成するスクリプトを記述する。\nbusyboxのbashで実行される。\n\"スクリプト名\":\"スクリプト実装\" のように記述する。"`
	DstTmp       string        `json:"tmp-dir"                  comment:"実行時に使用する作業用ディレクトリ。\n指定がないとき、起動時のカレントディレクトリ直下にtmpディレクトリを作成して使用する。"`
	LogPath      string        `json:"log"                      comment:"ログファイル出力先を指定する。この指定がないときログ出力しない。"`
	Silent       bool          `json:"quiet"                    comment:"標準出力と標準エラー出力に出力しない。"`
}

func (s *Setting) Print(args *Args) string { // {{{

	if s == nil {
		return "s is nil"
	}
	return fmt.Sprintf(`
	setting hjson path       : %q
	scripts                  : %q
	log                      : %q
	quiet                    : %v
	`,
		AbsPath(args.SettingPath),
		//constant                 : %q
		//s.Constant,
		s.Scripts,
		s.LogPath,
		s.Silent,
	)
} // }}}

func CreateEmptyHjson(args *Args) string { // {{{
	p := "./setting_empty.hjson"

	scripts := NewOrderedMapSS()
	scripts.Set("script_first.sh", `# ここにshellscriptを記述する。shebangは不要。
`)
	scripts.Set("script_second.sh", `# 同上
# スクリプトは必要なだけ追加できる。スクリプトの名前は任意だが重複は不可。
# 追加と同様に、スクリプトは不要なら削除できる。
`)
	scripts.Set("script_third.sh", `# 同上
`)

	var sb string
	s := Setting{Scripts: scripts}
	if true {
		// hjsonはOrderedMapSSをMarshalできないのでjsonとしてMarshalしてHjsonに変換する。
		jsonData, err := json.Marshal(s)
		dst := &bytes.Buffer{}
		if err := json.Compact(dst, jsonData); err != nil {
			panic(err)
		}
		jsonData = dst.Bytes()
		b, err := ConvertJSONdataToHJSON(jsonData)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		sb = string(b)
	} else {
		b, err := hjson.Marshal(s)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		sb = string(b)
	}

	// sb = insertStringAfter(sb, "constant: null", fmt.Sprintf(`constant:
	// '''
	// # ここに生成するshellscript全てで共通して使用したい定数を定義する。
	// foo: 42
	// '''
	// `))

	sbs := strings.Split(sb, "\n")

	for i, l := range sbs {
		sbs[i] = strings.TrimLeftFunc(l, unicode.IsSpace) // インデントを削除する。
	}

	sb = strings.Join(sbs, "\n")
	sb = sb[2:len(sb)-2] + "\n" // 最初と最後の{}を削除する。
	sb = RemoveEmptyLines(sb)

	sb = `##################################################################################################################################
#
# 設定ファイルの雛形。
#
# - このファイルの名前はsetting.hjsonに変更すること。他のパスや名前の場合、オプション'-s'で指定すること。
#
# - hjson形式。hjson は簡単に言えばコメントが書ける書きやすいjson。hjsonの書式の詳細は公式サイトを参照。 https://hjson.github.io/try.html
#     - ファイルパス文字列は\は/で記述すること。
#     - コメントには #と//と/**/の形式が使える。
#         - ファイルサーバー上のパスなどで、//で始まる場合、コメント扱いされるのを避けるために "" で括る必要がある。
#     - 文字列は上記の//を除いて引用符で括る必要はなく、そのまま書いてよい。
#         - ヒアドキュメントのように、文字列を複数行書きたいときは以下のように'''で括ることで記述できる。下記は例。
#           '''
#           ここに複数行記述する。
#           ここに複数行記述する。
#           '''
#
# - 不要な設定値は削除してよい。
# - 設定値がファイルパスの時空白があると誤動作する可能性がある。
# - 実際にはjsonに変換されるものと思ってよい。 サブコマンド'convertwを使用して、setting.hjsonをjsonに変換できる。
# - より具体的な例が見たいときは、 ./create_installer_iso.exe template のようにして追加の雛形を生成できる。
#   また、テスト用コマンド ./create_installer_iso.exe test を実行すると更に具体的な実行例が生成できるので参考に。
#
##################################################################################################################################


` + sb + "\n"

	sb = IndentOutermostBraces(sb)

	log.Printf(sb)

	WriteText(&p, &sb)
	return sb
} // }}}

//func OutputSettingTemplate(args *Args, embedFileList *EmbedFileList) {
//	LoadEmbedFiles() // 埋め込んだファイル群を一時ディレクトリにロード
//	Copy(embedFileList.TemplateSettingSimple, "./setting_simple_example_example.hjson", embedFileList)
//	Copy(embedFileList.TemplateSettingComplex, "./setting_complex_example_example.hjson", embedFileList)
//}

func NewSetting(args *Args, embedFileList *EmbedFileList) (*Setting, error) {
	LoadEmbedFiles() // 埋め込んだファイル群を一時ディレクトリにロード
	_ = os.RemoveAll(args.SettingPath)
	s := &Setting{}
	return ReadSetting(args, s)
}

func ReadSetting(args *Args, s *Setting) (*Setting, error) { // {{{

	//if !fileExists(args.SettingPath) {
	//	// 設定ファイルがなかった
	//	buf := new(bytes.Buffer)
	//	parser.WriteUsage(buf)
	//	fmt.Println(buf.String())
	//	panic(errors.Errorf("設定ファイル %v がありませんでした。\n", args.SettingPath))
	//}

	tmp := &Setting{}

	//log.Printf("setting : %v", args.SettingPath)

	b, err := os.ReadFile(args.SettingPath)
	if err != nil {
		return nil, err
	}
	if err := hjson.Unmarshal(b, tmp); err != nil {
		panic(errors.Errorf("%v", err))
	}

	tmp.DstTmp = filepath.ToSlash(filepath.Join(GetCurrentDir(), "tmp"))

	s = tmp

	// replaceConstant(s)
	//replaceVariable(s)

	s.Print(args)

	return s, nil
} // }}}

/*
func replaceConstant(s *Setting) { // {{{
	// 変数内の定数を置き換える。
	for ck, cv := range s.Constant {
		c := fmt.Sprintf("${%v}", ck)
		for k, v := range s.Variable {
			s.Variable[k] = strings.ReplaceAll(v, c, cv)
		}
	}

	// 以後定数参照するのが面倒なので変数に定数を追加
	for ck, cv := range s.Constant {
		s.Variable[ck] = cv
	}

} // }}}

func replaceVariable(s *Setting) { // {{{
	for k, v := range s.Variable {
		vk := fmt.Sprintf("${%v}", k)
		s.Dst = strings.ReplaceAll(s.Dst, vk, v)
		s.DstTmp = strings.ReplaceAll(s.DstTmp, vk, v)

		// s.SrcDiv       [][]string
		for i := 0; i < len(s.SrcDiv); i++ {
			for j := 0; j < len(s.SrcDiv[i]); j++ {
				s.SrcDiv[i][j] = strings.ReplaceAll(s.SrcDiv[i][j], vk, v)
			}
		}
		// s.SrcDivZip    [][]string
		for i := 0; i < len(s.SrcDivZip); i++ {
			for j := 0; j < len(s.SrcDivZip[i]); j++ {
				s.SrcDivZip[i][j] = strings.ReplaceAll(s.SrcDivZip[i][j], vk, v)
			}
		}
		for i, vv := range s.SrcCommon {
			s.SrcCommon[i] = strings.ReplaceAll(vv, vk, v)
		}
		for i, vv := range s.SrcCommonZip {
			s.SrcCommonZip[i] = strings.ReplaceAll(vv, vk, v)
		}
		for i, vv := range s.SrcOnlyDL {
			s.SrcOnlyDL[i] = strings.ReplaceAll(vv, vk, v)
		}
		for i, vv := range s.OutputIsoName {
			s.OutputIsoName[i] = strings.ReplaceAll(vv, vk, v)
		}
		for i, vv := range s.OutputIsoDir {
			s.OutputIsoDir[i] = strings.ReplaceAll(vv, vk, v)
		}
		for i, vv := range s.DiskLabels {
			s.DiskLabels[i] = strings.ReplaceAll(vv, vk, v)
		}
		s.ScriptForDst = strings.ReplaceAll(s.ScrirtForDst, vk, v)
		s.ScriptPost = strings.ReplaceAll(s.ScriptPost, vk, v)
		for i, vv := range s.ScriptForSpecificDisk {
			s.ScriptForSpecificDisk[i] = strings.ReplaceAll(vv, vk, v)
		}
		s.LogPath = strings.ReplaceAll(s.LogPath, kk, v)
	}
} // }}}
*/
