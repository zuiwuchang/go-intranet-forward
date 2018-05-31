#!/bin/bash
#Program:
#       自動編譯 google's Protocol Buffers 到 go 源碼
#History:
#       2018-03-09 king first release
#Email:
#       zuiwuchang@gmail.com

# 創建 輸出檔案夾
function mkOutDir(){
	if test -d $1 ;then
		return
	fi
	
	mkdir -p $1
	ok=$?
	if [[ $ok != 0 ]];then
		exit $ok
	fi
}
# 查找 pb 檔案
function findPB(){
	_f_pb_ctx=`pwd`
	cd $1 && files=`find *.proto -type f` && cd $_f_pb_ctx
	ok=$?
	if [[ $ok != 0 ]];then
		exit $ok
	fi
}
# 編譯 pd
function BuildPB(){
	#mkOutDir $2
	
	findPB $1
	echo "protoc -I $1 --go_out=$2" $files
	protoc -I $1 --go_out=$2 $files
	ok=$?
	if [[ $ok != 0 ]];then
		exit $ok
	fi
}

# 編譯 grpc
function BuildGRPC(){
	mkOutDir $2
	
	findPB $1
	echo "protoc -I $1 --go_out=plugins=grpc:$2" $files
	protoc -I $1 --go_out=plugins=grpc:$2 $files
	ok=$?
	if [[ $ok != 0 ]];then
		exit $ok
	fi
}

mkOutDir protocol/go/pb
# 編譯 pb
BuildPB protocol/pb protocol/go/pb

