# 毕昇 JDK 25.0.5 内网安装包（x64 + aarch64）

## 目录内容

```
bisheng-jdk-25.0.5-b16-linux-x64.tar.gz       # Intel / 海光（x86_64），本地文件，不入库
bisheng-jdk-25.0.5-b16-linux-x64.tar.gz.sha256
bisheng-jdk-25.0.5-b16-linux-aarch64.tar.gz   # 鲲鹏（aarch64），本地文件，不入库
bisheng-jdk-25.0.5-b16-linux-aarch64.tar.gz.sha256
parts/                                          # 两个 tar.gz 的 20MB 分片（入库，供内网拉取）
merge_parts.sh                                  # 分片合并 + SHA256 校验脚本
install_bisheng_jdk.sh                          # 一键安装脚本
download.sh                                     # 联网机刷新脚本（可选）
.gitignore                                      # 忽略原始整包，只提交分片
```

## 从 git 仓库取包（分片合并）

仓库里的安装包按 20MB 切片存放，拉取后先合并还原，再用于离线安装：

```bash
./merge_parts.sh          # 合并 parts/ 分片 → 还原两个 .tar.gz，并自动做 SHA256 校验
bash install_bisheng_jdk.sh --offline .
```

脚本行为：

| 动作 | 说明 |
| --- | --- |
| 分片检查 | 缺少分片、除末片外不是正好 20MB（漏片/截断）都会直接报错退出 |
| 合并 | 按 `part.00`、`part.01` … 顺序拼接，先写 `.tmp`，校验通过后才改名为正式包 |
| 校验 | 用同名 `.sha256` 比对 SHA256，失败则删除临时文件并不留下残缺包 |
| 幂等 | 已存在且校验通过的包会跳过，不会重复合并 |

> 手工逐个切分/合并可用 `split -b 20m` 与 `cat parts/<包名>.part.* > <包名>`，条数需与仓库中一致。

## 一条命令安装（推荐）

把本目录上传到内网文件服务器（例如 `http://10.0.0.10/jdk/`），目标机执行：

```bash
curl -fsSL http://10.0.0.10/jdk/install_bisheng_jdk.sh | sudo bash -s -- -u http://10.0.0.10/jdk
```

没有 `curl` 时：

```bash
wget -qO- http://10.0.0.10/jdk/install_bisheng_jdk.sh | sudo bash -s -- -u http://10.0.0.10/jdk
```

部署前可以把脚本第 23 行 `DEFAULT_FILE_SERVER` 改成文件服务器地址，之后命令里就不用再带 `-u`：

```bash
curl -fsSL http://10.0.0.10/jdk/install_bisheng_jdk.sh | sudo bash
```

## 脚本会做什么

| 步骤 | 说明 |
| --- | --- |
| 识别 CPU | `uname -m` → x86_64 选 x64 包，aarch64 选 aarch64 包；同时识别 Intel/海光/鲲鹏/飞腾型号，飞腾与海光给出灰度提示 |
| 下载 + 校验 | 从文件服务器下载对应架构的 tar.gz，用同名 `.sha256` 校验；自动探测 `wget` / `curl` |
| glibc 检查 | ≥ 2.18 正常；= 2.17 告警（如 CentOS 7.6），低于 2.17 中止 |
| 解压 | 安装到 `/opt/jdk`（`-d` 可改），建 `current` 软链指向实际目录 |
| 校验架构 | 读取 `bin/java` 的 ELF e_machine，防止文件服务器放错架构包 |
| 写环境变量 | root 写 `/etc/profile.d/bisheng-jdk.sh`；非 root 写 `~/.bashrc` |
| 打印结果 | `java -version`、vendor、`os.arch`，可直接复制进回归报告 |

重复执行幂等：已安装则跳过下载与解压，只刷新软链与环境变量。

## 其他用法

```bash
# 非 root 安装到 ~/jdk
bash <(curl -fsSL http://10.0.0.10/jdk/install_bisheng_jdk.sh) -d ~/jdk -u http://10.0.0.10/jdk

# 安装包已拷到本机，离线安装
bash install_bisheng_jdk.sh --offline /data/pkg

# 强制指定架构（默认自动识别）
bash install_bisheng_jdk.sh --offline . -d /opt/jdk -a aarch64

# 回滚（需先安装 25.0.4）
bash install_bisheng_jdk.sh --rollback

# 查看帮助
bash install_bisheng_jdk.sh -h
```

## 注意事项

1. **不要改文件名**：脚本按 `<url>/<包名>` 拼接，文件名与 `.sha256` 必须一一对应。
2. **解压后目录**：顶层目录名为 `bisheng-jdk-25.0.5`（不含构建号与架构），两台机器架构不同也不会冲突（`current` 软链指向各自目录）。
3. **CentOS 7.6**：glibc 2.17，脚本会告警但允许继续运行；官方不支持，长期建议迁移麒麟 V10。
4. **macOS 本机**：毕昇 JDK 只有 Linux 包，macOS 开发用 Oracle/Temurin JDK 25 替代，`JAVA_HOME` 切换即可。

## 校验 SHA256（上传前自检）

```bash
cd jdk_install
shasum -a 256 -c *.sha256
```
