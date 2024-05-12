package main

/*

func debug_output_line() {
	if developing {
		fmt.Println("------------------------------------")
	}
}

func Test(args *Args, s *Setting, embedFileList *EmbedFileList) {

	// {{{
	ans := `./output:
disk1.iso
disk2.iso
disk3.iso
iso_size.txt
tmp

./output/tmp:
a
a.7z
aaa
b
b.zip
bbb
c
c.rar
ccc
common
common.tar
common_nounzip.7z
create_iso_disk1.log
create_iso_disk1.sh
create_iso_disk2.log
create_iso_disk2.sh
create_iso_disk3.log
create_iso_disk3.sh
d
ddd
disk1
disk2
disk3
disk4
script_after_iso_create.log
script_after_iso_create.sh
script_for_all_disk_output
script_for_dst.log
script_for_dst.sh
script_for_specific_disk1.log
script_for_specific_disk1.sh
script_for_specific_disk2.log
script_for_specific_disk2.sh
script_for_specific_disk3.log
script_for_specific_disk3.sh
script_for_specific_disk4.log
script_for_specific_disk4.sh
setting.hjson

./output/tmp/a:
a_1
a_10
a_2
a_3
a_4
a_5
a_6
a_7
a_8
a_9

./output/tmp/b:
b_1
b_10
b_2
b_3
b_4
b_5
b_6
b_7
b_8
b_9

./output/tmp/c:
c_1
c_10
c_2
c_3
c_4
c_5
c_6
c_7
c_8
c_9

./output/tmp/common:
common_1
common_10
common_2
common_3
common_4
common_5
common_6
common_7
common_8
common_9

./output/tmp/d:
d_1
d_10
d_2
d_3
d_4
d_5
d_6
d_7
d_8
d_9

./output/tmp/disk1:
a_1
a_10
a_2
a_3
a_4
a_5
a_6
a_7
a_8
a_9
common_1
common_10
common_2
common_3
common_4
common_5
common_6
common_7
common_8
common_9
common_nounzip.7z
script_for_all_disk_output

./output/tmp/disk2:
b_1
b_10
b_2
b_3
b_4
b_5
b_6
b_7
b_8
b_9
script_for_all_disk_output

./output/tmp/disk3:
c_1
c_10
c_2
c_3
c_4
c_5
c_6
c_7
c_8
c_9
script_for_all_disk_output

./output/tmp/disk4:
d

./output/tmp/disk4/d:
d_1
d_10
d_2
d_3
d_4
d_5
d_6
d_7
d_8
d_9
`
	// }}}

	// テスト中にエラー終了したときは一時フォルダを削除しない。
	gDebugRetrainTmpDir = true

	// テストデータの展開
	ChangeDir(embedFileList.TmpDir)
	//src := GetRelativePath(embedFileList.TestData)
	dir := filepath.ToSlash(filepath.Join(GetCurrentDir(), "test data"))

	//runCommandByArrayWithoutPanic(embedFileList.BinBusybox, "bash", "rm", "-rf", "./test data")
	{
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "cp", "-rf", embedFileList.TestData, "./test data"), true, "")
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
		//fmt.Println("------------------------------------")
	}

	p := filepath.ToSlash(filepath.Join(dir, "test.sh"))
	script := "#!/bin/bash\n" + ReplaceCommand(fmt.Sprintf(`
set -ex
cd "$(dirname "$0")/test data"
# "%v/test data" で実行される。
pwd
ls -al
echo "### test : script-gen ###"
../../../create_installer_iso.exe script-gen -t
echo "### test : get ###"
../../../create_installer_iso.exe get -t
echo "### test : merge ###"
../../../create_installer_iso.exe merge -t
echo "### test : script-run ###"
../../../create_installer_iso.exe script-run -t
echo "### test : create ###"
../../../create_installer_iso.exe create  -t
pwd

# ディレクトリ構造をチェックする。
ls -R ./output | tee ls-R.log
`, dir), embedFileList)
	WriteText(&p, &script)
	if !args.Yes {
		if !prompter.YN("テストデータを生成しました。実行しますか。", true) {
			return
		}
		fmt.Println("テストを実行します。")
	} else {
		// --yesの時
		fmt.Println("テストデータを生成しました。実行しますか。 (y/n) [y]: テストを実行します。")
	}

	var copied_busybox string
	var ret string
	{
		Copy(embedFileList.TestData, filepath.ToSlash(filepath.Join(embedFileList.TmpDir, "test data")), embedFileList)

		// busyboxをコピーする。これは、busyboxの中で同じbusyboxを呼ぶとエラー終了する問題への対応
		// かつ、テスト失敗時にテスト用shellscriptを直で実行したいとき呼びやすくするため。
		copied_busybox = filepath.ToSlash(filepath.Join(embedFileList.TmpDir, "test data", "busybox.exe"))

		Copy(embedFileList.BinBusybox, copied_busybox, embedFileList)

		// テストを実行
		runCommandByArrayWithoutPanic(copied_busybox, "time", "bash", "-x", p)
		t := ChangeFilePathExt(p, ".log")
		//ret, err = runCommandOutputTextWithoutPanicByArray(t, embedFileList.BinBusybox, "-x", p)
		//ret, err = runCommandOutputTextWithoutPanicByArray(t, copied_busybox, "bash", "-x", p)
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(copied_busybox, "time", "bash", "-x", p), true, t)

		debug_output_line()

		if !IsExist(t) {
			log.Printf("test結果のlogが見つかりませんでした。: %#v", t)
		}
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
	}

	debug_output_line()

	//log.Printf("pwd: %v", GetCurrentDir())
	ret = GetText_("./test data/test data/ls-R.log")

	if ans == ret {
		fmt.Println("OK!")
		if args.Yes || prompter.YN(fmt.Sprintf("テストの実行により一時ファイルが生成されています。: %v\nデータを削除しますか。", embedFileList.TmpDir), true) {
			os.RemoveAll(embedFileList.TmpDir)
			gDebugRetrainTmpDir = false
		}
	} else {
		fmt.Println("NG!!!")
	}
}
*/
