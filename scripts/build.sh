
function logInfo() {
    echo `date "+%F %T" ` "INFO:" $@ 1>&2
}
function logError() {
    echo `date "+%F %T" ` "ERROR:" $@ 1>&2
}
function logWarn() {
    echo `date "+%F %T" ` "WARN:" $@ 1>&2
}
function goBuild(){
    logInfo "获取版本"
    version=$(go run cmd/stackcrud.go -v |awk '{print $3}')
    if [[ -z $version ]] || [[ "${version}" == "" ]]; then
        exit 1
    fi
    mkdir -p dist
    logInfo "开始编译, 版本: ${version}"
    go build  -ldflags "-X main.Version=${version} -s -w" -o dist/ cmd/stackcrud.go
    if [[ $? -ne 0 ]]; then
        logError "编译失败"
        exit 1
    fi
    logInfo "编译成功"
    which upx > /dev/null 2>&1
    if [[ $? -eq 0 ]]; then
        logInfo "检测到工具 upx, 压缩可执行文件"
        upx -q dist/stackcrud > /dev/null
    else
        logWarn "upx未安装, 不压缩可执行文件"
    fi
}

function rpmBuild() {
    logInfo "构建rpm包"
    local buldingSpec=/tmp/stackcrud.spec

    rm -rf ${buldingSpec}
    cp release/stackcrud.spec ${buldingSpec} || exit 1
    local buildVersion=$(./dist/stackcrud -v |awk '{print $3}')

    sed -i "s|VERSION|${buildVersion}|g" ${buldingSpec}
    logInfo "版本: $(awk '/^Version/{print $2}' ${buldingSpec})"

    mkdir -p /root/rpmbuild/SOURCES
    cp dist/stackcrud etc/stackcrud-template.yaml /root/rpmbuild/SOURCES || exit 1
    rpmbuild -bb ${buldingSpec}

    ls -1 /root/rpmbuild/RPMS/x86_64/stackcrud-*.rpm |while read line
    do
        local rpmName=$(basename ${line})
        rm -rf dist/$line
        mv ${line} dist
    done

    rm -rf ${buldingSpec}
}

function main(){
    local buildRpm=false
    while [[ true ]]
    do
        case "$1" in
         --rpm)
            buildRpm=true
            shift
            ;;
        *)
            if [[ -z ${1} ]]; then
                break
            else
                echo "ERROR: invalid arg $1";
                exit 1;
            fi
            ;;
        esac
    done
    if [[ ${buildRpm} == true ]]; then
        rpmBuild
    else
        goBuild
    fi
}
main $*
