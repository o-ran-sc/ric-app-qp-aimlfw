#==================================================================================
#      Copyright (c) 2022 Samsung Electronics Co., Ltd. All Rights Reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#
#==================================================================================
FROM nexus3.o-ran-sc.org:10002/o-ran-sc/bldr-ubuntu20-c-go:1.1.0 as builder

RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr_4.7.0_amd64.deb/download.deb && dpkg -i rmr_4.7.0_amd64.deb && rm -rf rmr_4.7.0_amd64.deb
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr-dev_4.7.0_amd64.deb/download.deb && dpkg -i rmr-dev_4.7.0_amd64.deb && rm -rf rmr-dev_4.7.0_amd64.deb

# Install necessary packages
RUN apt-get update && apt-get install -y \
    build-essential \
    && apt-get clean

ENV LD_LIBRARY_PATH /usr/local/lib
ENV PATH $PATH:/usr/local/bin:$GOPATH/bin

RUN mkdir /opt/qoe-aiml-assist
WORKDIR /opt/qoe-aiml-assist

ENV GO111MODULE=on GO_ENABLED=0 GOOS=linux

COPY . .

RUN go mod tidy -compat=1.17 && go mod vendor && go build -o build/qoe-aiml-assist

# for unittest
RUN sed -r "s/^(::1.*)/#\1/" /etc/hosts > /etc/hosts.new \
    && cat /etc/hosts.new > /etc/hosts \
    && cat /etc/hosts \
    && go test -v ./influx ./control -test.coverprofile /tmp/qp_cover.out \
    && go tool cover -html=/tmp/qp_cover.out -o /tmp/qp_cover.html

FROM ubuntu:20.04

ENV CFG_FILE=config/config-file.json
ENV RMR_SEED_RT=config/uta_rtg.rt

RUN apt update && apt install -y \
    iputils-ping \
    net-tools \
    curl \
    tcpdump \
    sudo \ 
    ca-certificates 

COPY --from=builder /opt/qoe-aiml-assist/build/qoe-aiml-assist .
COPY --from=builder /usr/local/include /usr/local/include
COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /opt/qoe-aiml-assist/config/* /opt/ric/config/
COPY --from=builder /tmp/qp_cover.* /tmp/

RUN ldconfig
