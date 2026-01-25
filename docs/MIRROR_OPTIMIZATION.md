# 镜像源优化说明

本文档说明了项目中使用的各种镜像源优化配置，以加速在中国大陆的构建和部署。

## 已优化的镜像源

### 1. Go 模块代理

所有 Go 服务的 Dockerfile 都配置了阿里云 Go 代理：

```dockerfile
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
ENV GO111MODULE=on
```

**影响的服务：**
- auth-service
- cmdb-service
- ticket-service
- alert-service
- notification-service
- report-service
- audit-service

### 2. Alpine Linux 软件包镜像

所有使用 Alpine Linux 的镜像都配置了阿里云镜像源：

```dockerfile
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
```

**影响的镜像：**
- 所有 Go 服务的 builder 和 runtime 阶段
- frontend 的 builder 和 runtime 阶段
- nginx (api-gateway)

### 3. npm 镜像源

前端 Dockerfile 配置了淘宝 npm 镜像：

```dockerfile
RUN npm config set registry https://registry.npmmirror.com && \
    npm install
```

**影响的服务：**
- frontend

## Docker Hub 镜像加速（可选）

如果需要加速 Docker Hub 镜像拉取，可以配置 Docker daemon：

### Linux 系统

编辑 `/etc/docker/daemon.json`：

```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
```

重启 Docker 服务：

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### Docker Desktop (Windows/Mac)

1. 打开 Docker Desktop 设置
2. 进入 "Docker Engine" 选项卡
3. 添加以下配置：

```json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
```

4. 点击 "Apply & Restart"

## 构建速度对比

使用镜像源优化后，预计构建速度提升：

| 阶段 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| Go 依赖下载 | 30-60s | 5-10s | 5-6x |
| Alpine 包安装 | 10-20s | 2-5s | 3-4x |
| npm 依赖安装 | 60-120s | 15-30s | 4x |

## 验证配置

### 验证 Go 代理

```bash
docker compose build --no-cache auth-service 2>&1 | grep GOPROXY
```

应该看到：`ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct`

### 验证 Alpine 镜像源

```bash
docker compose build --no-cache auth-service 2>&1 | grep mirrors.aliyun.com
```

应该看到镜像源替换的输出。

### 验证 npm 镜像源

```bash
docker compose build --no-cache frontend 2>&1 | grep registry.npmmirror.com
```

应该看到 npm 配置的输出。

## 故障排除

### 如果镜像源不可用

如果某个镜像源不可用，可以替换为其他镜像源：

**Go 代理备选：**
- `https://goproxy.cn`
- `https://goproxy.io`
- `https://mirrors.tencent.com/go/`

**Alpine 镜像备选：**
- `mirrors.tuna.tsinghua.edu.cn`
- `mirrors.ustc.edu.cn`

**npm 镜像备选：**
- `https://registry.npm.taobao.org` (旧版)
- `https://registry.npmmirror.com` (新版，推荐)

## 更新日期

最后更新：2026-01-25
