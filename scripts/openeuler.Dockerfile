# https://gitee.com/openeuler/openeuler-docker-images
FROM hub.oepkgs.net/openeuler/openeuler:20.03-lts

RUN mkdir /etc/yum.repos.d/backup && mv  /etc/yum.repos.d/*.repo /etc/yum.repos.d/backup

COPY openEuler-20.03-lts.repo /etc/yum.repos.d
# Install required packages
RUN yum install -y wget git rpm-build rpmdevtools -y which make gcc glibc-devel libvirt-devel
# Install upx
RUN yum install -y https://mirrors.aliyun.com/epel/7/x86_64/Packages/u/ucl-1.03-24.el7.x86_64.rpm
RUN yum install -y https://mirrors.aliyun.com/epel/7/x86_64/Packages/u/upx-3.96-9.el7.x86_64.rpm
# Install golang
RUN wget -q https://golang.google.cn/dl/go1.24.1.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
RUN cp /usr/local/go/bin/* /usr/bin/

RUN export PATH=/usr/local/go/bin:$PATH \
    && which go \
    && go env -w GOROOT='/usr/local/go' \
    && go env -w GO111MODULE="on" \
    && go env -w GOPROXY="https://mirrors.aliyun.com/goproxy/,direct" \
