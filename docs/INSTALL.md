Installation instructions for ubuntu 17.10

# get dependencies


## get docker
docker is need for crosscompiling ->
[docker ubuntu installation manual](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)


<!-- ## get go-qml dependencies
```
sudo add-apt-repository ppa:ubuntu-sdk-team/ppa
sudo apt-get update
sudo apt-get install qtdeclarative5-dev qtbase5-private-dev qtdeclarative5-private-dev libqt5opengl5-dev qtdeclarative5-qtquick2-plugin
sudo ln -s /usr/include/x86_64-linux-gnu/qt5/QtCore/5.9.1/QtCore /usr/include/
```


## get go lang

```
# This will give you the latest version of go
snap install --classic go -->

```
##set $GOPATH
## get go dependencies
```
go get -v -d github.com/sirupsen/logrus
go get -v -d github.com/godbus/dbus
go get -v -d github.com/dustin/go-humanize
go get -v -d github.com/godbus/dbus
go get -v -d github.com/gosexy/gettext
go get -v -d github.com/nanu-c/textsecure
go get -v -d github.com/jmoiron/sqlx
go get -v -d github.com/mattn/go-sqlite3
go get -v -d github.com/ttacon/libphonenumber
go get -v -d github.com/snapcore/snapd/osutil/sys
go get -v -d github.com/morph027/textsecure
go get -v -d gopkg.in/yaml.v2
go get -v -d bitbucket.org/llg/vcard
```
