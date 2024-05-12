# gen_sh

複数のshellscriptを修正して実行といった作業の自動化補助ツール。
この設定ファイルを使用して、複数個のshellscriptを、設定ファイル内の値を用いて自動生成して、一括または個別実行できる。  

```sh
$ ./gen_sh.exe 
gen_sh.exe version x.x.x.

Description:
  複数のshellscriptを修正して実行といった作業の自動化補助ツール。
  この設定ファイルを使用して、複数個のshellscriptを、設定ファイル内の値を用いて自動生成して、一括または個別実行できる。

Usage: gen_sh.exe [--check-setting] [--setting FILE] [--cd CD] [--show-readme] [--logfile FILE] [--quiet] [--retrain-tmp] [--version] [--help] <command> [<args>]

Options:
  --check-setting        設定ファイルを読みこんでどのように解釈されたかを出力する。
  --setting FILE, -s FILE
                         使用したい設定ファイルパス指定。指定がなければ./setting.hjsonを使用する。 [default: ./setting.hjson]
  --cd CD, -c CD         カレントディレクトリを指定のディレクトリに変更してから実行する。
  --show-readme, -r      readme.mdを出力する。
  --logfile FILE, -l FILE
                         ログファイル出力先を指定する。設定ファイルで指定されておらず、この指定がないときログ出力しない。
  --quiet, -q            標準出力と標準エラー出力に出力しない。
  --retrain-tmp, -t      実行時に生成した一時ディレクトリを削除しない。
                         script-genなどで生成したshellscriptを編集した後に、一時ファイル内のbusyboxやmkisofsなどを利用できる。
  --version, -v          バージョン番号を出力する。サブコマンドversionと同じ。
  --help
  --help, -h             display this help and exit

Commands:
  init                   空の設定ファイルを生成する。
  convert                設定ファイルをjsonに変換する。スクリプト内でjsonをjqで参照する、のような用途を想定。
  generate               設定ファイルに記述したスクリプトからshellscirptを生成する。既にあれば上書きされるので注意。
  run                    生成されたスクリプトを実行する。引数に実行する設定ファイルに記載したスクリプト名を指定する。
  run-all                適切な設定ファイルがある前提で、get,merge,script-gen,script-run,createを一括で実行する。
  code                   このプログラムのソースコードを出力する。golangのコンパイラは https://go.dev/dl/ からダウンロードできる。
  version                バージョン番号を出力する。-vと同じ。

```


```hjson
```
