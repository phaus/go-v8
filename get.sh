#!sh

if [ $# -eq 0 ]; then
	echo >&2 "useage: ./get.sh the_install_dir"
	exit
fi

install_dir=$1

mkdir $install_dir

cd $install_dir
export GOPATH=`pwd`

mkdir -p bin pkg src/github.com/idada/v8.go
cd src/github.com/idada/v8.go

git init
git remote add origin https://github.com/idada/v8.go
git pull origin master

./install.sh
