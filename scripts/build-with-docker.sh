set -u

realPath=$(realpath scripts/build-with-docker.sh)
scriptPath=$(dirname ${realPath})
projectPath=$(dirname ${scriptPath})

function main(){
    which docker > /dev/null || exit 1
    docker ps > /dev/null || exit 1

    echo "INFO" "项目路径" $projectPath

    echo "INFO" "========== 构建基础镜像 =========="
    cd ${scriptPath}
    time docker build --tag skyman-openeuler-builder \
        -f openeuler.Dockerfile \
        ./
    if [[ $? -ne 0 ]]; then
        echo "ERROR" "基础镜像构建失败"
        exit 1
    fi
    cd -
    echo "INFO" "========== 开始编译 =========="
    containerName=skyman-openeuler-builder
    docker container inspect ${containerName} > /dev/null 2>&1
    if [[ $? -eq 0 ]]; then
        echo "INFO" "容器已存在, 启动 ..."
        docker start ${containerName}
    else
        docker run -itd \
            --name ${containerName} \
            -v $(pwd):/mnt \
            skyman-openeuler-builder \
            sh -c 'cd /mnt && make '
    fi
    echo "INFO" "等待容器运行结束 ..."
    time docker logs --follow ${containerName}
}

main $*
