# stackcrud

Goalng Openstack client

## Oveview

```
$ stackcrud
Golang Openstack Client

Usage:
  stackcurd [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  compute
  flavor
  help        Help about any command
  hypervisor
  image
  server
  volume

Flags:
  -c, --conf stringArray   配置文件 (default [etc/stackcrud.yaml,/etc/stackcrud/stackcrud.yaml])
  -d, --debug              显示Debug信息
  -h, --help               help for stackcurd
  -v, --version            version for stackcurd

Use "stackcurd [command] --help" for more information about a command.
```

### 构建

1. 编译
   
   ```bash
   sh scripts/build.sh
   ```
   
   > 输出目录: ./dist

2. 构建 rpm 包（使用podman或docker）
   
   ```bash
   sh scripts/build-with-docker.sh
   ```
   
   > 输出目录: ./dist

## 安装

```bash
rpm -ivh dist/stackcrud-<版本>-1.x86_64.rpm
```
