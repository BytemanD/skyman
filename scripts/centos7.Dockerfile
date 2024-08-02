FROM 93s63uis.mirror.aliyuncs.com/library/centos:7.8.2003 as Skylight-Centos7-Base

RUN mkdir /etc/yum.repos.d/backup && mv /etc/yum.repos.d/*.repo /etc/yum.repos.d/backup
RUN curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo

# Install required packages
RUN yum install -y wget git rpm-build rpmdevtools -y which make gcc glibc-devel libvirt-devel
# Install upx
RUN yum install -y https://mirrors.aliyun.com/epel/7/x86_64/Packages/u/ucl-1.03-24.el7.x86_64.rpm
RUN yum install -y https://mirrors.aliyun.com/epel/7/x86_64/Packages/u/upx-3.96-9.el7.x86_64.rpm
# Install golang
RUN wget -q https://dl.google.com/go/go1.21.4.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
RUN cp /usr/local/go/bin/* /usr/bin/

RUN echo 'export PATH=/usr/local/go/bin:$PATH' >> $HOME/.bashrc
RUN source $HOME/.bashrc && /usr/local/go/bin/go version

# Build Skylight
FROM Skylight-Centos7-Base as Skylight-Centos7-Builder

RUN source $HOME/.bashrc \
    && cd /root/skyman \
    && go env -w GO111MODULE="on" \
    && go env -w GOPROXY="https://mirrors.aliyun.com/goproxy/,direct" \
    && go mod download

# NOTE:In order not to use caching
ARG DATE
RUN echo ${DATE}

RUN source $HOME/.bashrc \
    && go env -w GO111MODULE="on" \
    && go env -w GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"

RUN cd /root/skyman \
    && source $HOME/.bashrc \
    && make build build-rpm
