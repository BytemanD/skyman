
function logInfo() {
    echo `date "+%F %T" ` "INFO:" $@ 1>&2
}

function main(){
    logInfo "Get version"
    version=$(go run cmd/stackcrud.go -v |awk '{print $3}')
    if [[ -z $version ]] || [[ "${version}" == "" ]]; then
        exit 1
    fi
    logInfo "Start to build with option main.Version=${version}"
    mkdir -p dist
    go build  -ldflags "-X main.Version=${version}" -o dist/ cmd/stackcrud.go
    logInfo "Build success"
}
main
