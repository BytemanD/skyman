# skyman

Goalng OpenStack Client

## Oveview

```
$ skyman
Golang Openstack Client

Usage:
  skyman [command]

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
  -c, --conf stringArray   配置文件 (default [etc/skyman.yaml,/etc/skyman/skyman.yaml])
  -d, --debug              显示Debug信息
  -h, --help               help for skyman
  -v, --version            version for skyman

Use "skyman [command] --help" for more information about a command.
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
rpm -ivh dist/skyman-<版本>-1.x86_64.rpm
```
## 设置语言

```bash
export SKYMAN_LANG=zh_CN
```

## 支持的命令

```bash
skyman aggregate list/show
skyman az list
skyman compute service list/enable/disable/up/down
skyman console log/url
skyman action list/show
skyman flavor list/create/delete/copy
skyman hypervisor list
skyman keypair list
skyman migration list
skyman server list/show/create/delete/prune
skyman server stop/start/reboot/pause/unpause
skyman server shelve/unshelve/suspend/resume
skyman server resize/migrate/rebuild
skyman server interface list/attach-port/attach-net/detach
skyman server volume list/attach/detach

skyman image list

skyman volume list/show/delete/prune

skyman network list
skyman router list
skyman port list
...
```

