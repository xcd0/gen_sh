package main

/*

func Create(args *Args, s *Setting, embedFileList *EmbedFileList) error {
	// "github.com/kdomanski/iso9660"
	// で作ったISOはwindowsで正しく動作しない。
	// issueで１年前から上がっているが何ら対応がないので、
	// kdomanski/iso9660は諦める。

	//log.Printf("Create")

	if args.Create != nil {
		if len(args.Create.SrcDir) != 0 && len(args.Create.DiskLabels) != 0 {
			if false {
				log.Printf("args.Create.SrcDir     : %v", args.Create.SrcDir)
				log.Printf("args.Create.DiskLabels : %v", args.Create.DiskLabels)
				log.Printf("args.Create.Dst        : %v", args.Create.Dst)
			}

			args.Create.Dst = getFileNameWithoutExt(args.Create.Dst)
			args.Create.Dst = AbsPath(args.Create.Dst + ".iso")
			args.Create.SrcDir = AbsPath(args.Create.SrcDir)

			CreateISO(
				args.Create.DiskLabels,
				args.Create.Dst+".iso",
				args.Create.SrcDir,
				s, embedFileList)
			return nil
		} else if s == nil {
			log.Fatalf("Setting file could not load. : %s", args.SettingPath)
			return nil
		}
	}

	// 設定ファイルに従ってISOを生成
	for i := 0; i < len(s.DiskLabels); i++ {
		// 出力先/ISOファイル名/をISOに格納するして
		// 出力先/ISOファイル名.isoに出力する
		s.OutputIsoName[i] = getFileNameWithoutExt(s.OutputIsoName[i])

		output_iso := s.OutputIsoName[i] + ".iso"
		src_dir := s.OutputIsoDir[i]

		if !false {
			log.Printf("pwd : %v", GetCurrentDir())
			log.Printf("Dst : %v", s.DstTmp)
			log.Printf("dst : %v", output_iso)
			log.Printf("src : %v", src_dir)
		}

		CreateISO(
			s.DiskLabels[i],
			output_iso, src_dir,
			s, embedFileList)
	}

	// ISO作成後に実行するスクリプトを実行する。
	if len(s.ScriptPost) != 0 {

		pwd := s.DstTmp
		log.Printf("RunScript: ChangeDir(%#v)", pwd)
		ChangeDir(pwd)

		p := "script_after_iso_create.sh"
		t := ChangeFilePathExt(p, ".log")
		_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "bash", "-ex", p), true, t)
		if err != nil {
			panic(errors.Errorf("%v", err))
		}
	}

	return nil
}

func CreateISO(label, output_iso, src_dir string, s *Setting, embedFileList *EmbedFileList) {

	pwd := s.DstTmp
	log.Printf("RunScript: ChangeDir(%#v)", pwd)
	ChangeDir(pwd)
	//pwd := GetCurrentDir()

	// in := "./" + filepath.ToSlash(GetRelativePath(AbsPath(src_dir)))
	// out := "./" + filepath.ToSlash(GetRelativePath(AbsPath(output_iso)))
	in := "./" + filepath.ToSlash(GetRelativePath(src_dir))
	out := "./" + filepath.ToSlash(GetRelativePath(output_iso))

	if false {
		log.Printf("src_dir : %v", src_dir)
		log.Printf("output_iso : %v", output_iso)
		log.Printf("in  : %v", in)
		log.Printf("out : %v", out)
	}

	//  ISO9660のファイルシステムCDFSのISOを作る場合(windows XPぐらい古いやつ用)
	//cmd := fmt.Sprintf("%s -D -l -J -r -input-charset=utf-8 -allow-leading-dots -V %s -o %s %s", embedFileList.BinMkisofs, label, out, in)
	// UDFのファイルシステムのISOを作る場合(こっちでよい)
	//cmd := fmt.Sprintf("%s bash %s -UDF -V %s -o %s %s", embedFileList.BinBusybox, embedFileList.BinMkisofs, label, out, in)
	//log.Printf(cmd)
	//RunCommand(cmd)

	option := ""
	if len(s.IsoFilesystem) == 0 {
		option = "-UDF -r -input-charset=utf-8"
		// -rをつけないとディレクトリの深さの最大値が8になる模様。以下のような警告が出る。max is 6と出ているが8までっぽい。
		// mkisofs: Directories too deep for '/hoge/深いパス' (7) max is 6; ignored - continuing.
		// mkisofs: To include the complete directory tree,
		// mkisofs: use Rock Ridge extensions via -R or -r,
		// mkisofs: or allow deep ISO9660 directory nesting via -D.
	} else if s.IsoFilesystem[0] == '-' {
		option = s.IsoFilesystem
	} else if strings.EqualFold(s.IsoFilesystem, "UDF") {
		option = "-UDF -r -input-charset=utf-8"
	} else if strings.EqualFold(s.IsoFilesystem, "CDFS") {
		option = "-D -l -J -r -input-charset=utf-8 -allow-leading-dots"
	} else {
		option = "-UDF -r -input-charset=utf-8"
	}

	script := fmt.Sprintf(`#!/bin/bash

# 実行ファイルに埋め込まれているバイナリのあるディレクトリにパスを通す。
export PATH="%v:$PATH"

# %#v で実行される。

time mkisofs %v -V %#v -o %#v %#v
`,
		filepath.ToSlash(filepath.Join(embedFileList.TmpDir, "files/bin")), // ここにPATHを通す。
		pwd,
		option,
		label,
		out,
		in,
	)

	p := fmt.Sprintf("create_iso_%v.sh", filepath.Base(getFileNameWithoutExt(output_iso)))
	t := ChangeFilePathExt(p, ".log")
	WriteText(&p, &script)
	log.Printf("mkisofs script : %#v\n%v", p, script)
	_, _, _, err := runCommandOutputRealtimeWithTee(exec.Command(embedFileList.BinBusybox, "bash", "-ex", p), true, t)

	if err != nil {
		panic(errors.Errorf("%v", err))
	}
}
*/
