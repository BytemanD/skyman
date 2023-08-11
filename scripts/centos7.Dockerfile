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
RUN yum install -y which

# Install upx
RUN wget -q http://rpmfind.net/linux/epel/7/x86_64/Packages/u/ucl-1.03-24.el7.x86_64.rpm
RUN wget -q http://rpmfind.net/linux/epel/7/x86_64/Packages/u/upx-3.96-9.el7.x86_64.rpm
RUN rpm -ivh ucl-1.03-24.el7.x86_64.rpm upx-3.96-9.el7.x86_64.rpm

# Build project
FROM Stackcrud-Centos7-Base as Stackcrud-Centos7-Builder

# In order not to use caching
ARG DATE

RUN echo ${DATE}
RUN go env -w GO111MODULE="on" \
    && go env -w GOPROXY="https://goproxy.cn,direct"
RUN cd /root/stackcrud && sh scripts/build.sh
RUN cd /root/stackcrud && sh scripts/build.sh --rpm
