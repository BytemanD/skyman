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

1. 本地编译
   
   ```bash
   make
   ```

2. 通过docker 容器编译
   
   ```bash
   sh scripts/build-with-docker.sh
   ```
   
   保存目录: ./dist， 构建完成后，当前目录下将同时生成二进制文件和rpm 包

## 安装&更新

> 系统要求：>= openeuler 20.03

1. 直接使用二进制文件
   
   - 先拷贝 skyman 文件到节点任意目录下
   
   - 然后增加执行权限:`chmod u+x skyman` 
   
   - 导入 openstack 环境变量后，执行 ./skyman 命令

2. 安装rpm 包
   
   - 先安装rpm包： rpm -iUvh dist/skyman-<版本>-1.x86_64.rpm
   
   - 导入 openstack 环境变量后，执行 skyman 命令
     
     

## 设置

配置文件  /etc/skyman/clouds.yaml

1. 设置语言
   修改配置文件
   
   ```yaml
   language: en_US
   ```
   
   或者执行 `export SKYMAN_LANG=zh_CN`

2. 添加云环境
   
   ```yaml
   # ...
   clouds:
     mydev1:
       region_name: RegionOne
       auth:
         auth_url: <认证地址>
         project_domain_id: default
         user_domain_id: default
         project_name: <项目名>
         username: <用户名>
         password: <密码>
      mydev2:
         ...
   ```

3. 设置默认的云环境名称
   
   ```yaml
   cloud: mydev1
   ```
   
   或者执行执行命令是添加参数 `--cloud mydev1`
   
   

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

### 工具

- 并发挂载网卡（根据指定的network ， 自动创建port, 然后并行挂载到实例上）
  
  ```bash
  skyman tool server add-interfaces  <实例ID> <网络 UUID>
  ```
  
  > 可选项
  > --nums <挂载数量，默认为1>
  > --use-net-id 使用net id 挂载，而不是 port id
  
  

- 并发卸载网卡（查询实例的所有网卡 ，然后按照网卡列表，取最后N个网卡，并行卸载）yaml
  
  ```bash
  skyman tool server remove-interfaces <实例ID> 
  ```
  
  > 可选项:
  > --nums <卸载数量，默认为1>
  > --clean  卸载后，自动删除port
  
  

- **并发挂载卷（根据指定的network ， 自动创建port, 然后并行挂载到实例上）**
  
  ```
  skyman tool server add-interfaces  <实例ID> <网络 UUID>
  ```
  
  > 可选项:
  > --nums <挂载数量，默认为1>
  > --size 卷大小--type 卷类型

- 并发卸载网卡（查询实例的所有卷 ，然后按照卷列表，取最后N个数据卷，并行卸载）
  
  ```bash
  skyman tool server remove-volumes <实例ID>
  ```
  
  > 可选项:
  > --nums <卸载数量，默认为1>
  > --clean 卸载后，自动删除卷
