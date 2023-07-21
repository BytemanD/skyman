FROM 93s63uis.mirror.aliyuncs.com/library/centos:7.8.2003 as Stackcrud-Centos7-Base

# Install golang
RUN yum install -y wget
RUN wget -q https://golang.google.cn/dl/go1.19.11.linux-amd64.tar.gz
RUN tar -xzf go1.19.11.linux-amd64.tar.gz -C /usr/local/
RUN cp /usr/local/go/bin/* /usr/bin/
RUN go version

# Install required packages
RUN yum install -y git
RUN yum install -y rpm-build rpmdevtools

# Build project
FROM Stackcrud-Centos7-Base as Stackcrud-Centos7-Builder

# In order not to use caching
ARG DATE

RUN echo ${DATE}
RUN go env -w GO111MODULE="on" \
    && go env -w GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"
RUN cd /root/stackcrud && sh scripts/build.sh
RUN cd /root/stackcrud && sh scripts/build.sh --rpm
