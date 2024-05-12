all:
	cat ./makefile
run:
	make -C ./src run
get:
	make -C ./src get
fmt:
	make -C ./src fmt
prebuild:
	./files/bin/busybox64u.exe bash -x ./prebuild.sh
build:
	make -C ./src build
release:
	make -C ./src release
upx:
	upx --lzma ./create_installer_iso.exe
template:
	./files/bin/busybox64u.exe time ./create_installer_iso.exe template
test:
	./files/bin/busybox64u.exe time ./create_installer_iso.exe test -y -t
version:
	make -C ./src version | sed -e 's;^;> ;'

