# Simple usage with a mounted data directory:
# > docker build -t bttcprotocal/delivery:<tag> .
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.deliveryd:/root/.deliveryd bttcprotocal/delivery:<tag> deliveryd init

# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:latest

# update available packages
RUN apt-get update -y && apt-get upgrade -y && apt install build-essential -y

# create go src directory and clone delivery
RUN mkdir -p /root/delivery

# change work directory
WORKDIR /root/delivery

# download mod
ADD go.mod /root/delivery/
ADD go.sum /root/delivery/
RUN go mod download

# add code to docker instance
ADD . /root/delivery/

# GOBIN required for go install
ENV GOBIN $GOPATH/bin

# run build
RUN make install

# add volumes
VOLUME [ "/root/.deliveryd", "./logs" ]

# expose ports
EXPOSE 1317 26656 26657
