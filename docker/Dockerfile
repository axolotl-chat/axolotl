FROM clickable/ubuntu-sdk:16.04-armhf

RUN apt update

RUN apt install -y libglib2.0-dev libgdk-pixbuf2.0-dev libcairo2-dev
RUN apt install -y librsvg2-dev

RUN add-apt-repository ppa:longsleep/golang-backports
RUN apt update
RUN apt-get install -y golang-go
RUN rm  /usr/local/go/bin/go && ln -s /usr/bin/go /usr/local/go/bin/go
RUN apt install -y libzbar-dev:armhf
