# GitHub Actions Docker 镜像自动构建指南

本指南介绍如何使用 GitHub Actions 自动构建和发布 Docker 镜像。

## 概述

GitHub Actions 工作流会自动：
- 构建多架构 Docker 镜像 (amd64 + arm64)
- 推送到 GitHub Container Registry (GHCR)
- 可选推送到 Docker Hub
- 支持版本标签管理
- 提供构建缓存加速

## 配置步骤

### 1. 启用 GitHub Packages

GitHub Container Registry (GHCR) 默认可用，无需额外配置。

镜像将发布到：
```
ghcr.io/<username>/gp-service:latest
ghcr.io/<username>/gp:latest
```

### 2. 配置 Docker Hub（可选）

如果要同时推送到 Docker Hub：

**步骤 1：创建 Docker Hub Access Token**
1. 登录 [Docker Hub](https://hub.docker.com/)
2. 进入 Account Settings → Security
3. 点击 "New Access Token"
4. 创建名为 `github-actions` 的 token
5. 复制生成的 token

**步骤 2：在 GitHub 仓库配置 Secrets**
1. 打开仓库 Settings → Secrets and variables → Actions
2. 点击 "New repository secret"
3. 添加以下 secrets：
   - `DOCKERHUB_USERNAME`: 你的 Docker Hub 用户名
   - `DOCKERHUB_TOKEN`: 上一步生成的 token

### 3. 工作流触发条件

工作流会在以下情况自动运行：

**自动触发：**
- 推送到 `main` 或 `develop` 分支
- 创建版本标签 (如 `v1.0.0`)
- 创建或更新 Pull Request

**手动触发：**
- 进入 Actions 标签页
- 选择 "Docker Build and Push" 工作流
- 点击 "Run workflow"

## 使用镜像

### 从 GitHub Container Registry 拉取

```bash
# 拉取最新版本
docker pull ghcr.io/<username>/gp-service:latest

# 拉取特定版本
docker pull ghcr.io/<username>/gp-service:v1.0.0

# 拉取特定 commit
docker pull ghcr.io/<username>/gp-service:main-abc1234
```

### 从 Docker Hub 拉取

```bash
# 拉取最新版本
docker pull <username>/alpha-detector-service:latest

# 拉取特定版本
docker pull <username>/alpha-detector-service:v1.0.0
```

### 公开镜像访问

**GHCR 镜像公开步骤：**
1. 进入 GitHub 个人主页 → Packages
2. 找到你的镜像包
3. 点击 Package settings
4. 在 "Danger Zone" 中选择 "Change visibility"
5. 设置为 Public

## 镜像标签策略

工作流会自动生成以下标签：

| 触发条件 | 生成的标签 | 示例 |
|---------|-----------|------|
| 推送到 main | `latest` | `latest` |
| 推送到 develop | `develop` | `develop` |
| 推送到其他分支 | `<branch>` | `feature-xyz` |
| 创建版本标签 | `v1.0.0`, `v1.0`, `v1` | `v1.2.3` |
| 任何提交 | `<branch>-<sha>` | `main-abc1234` |

## 多架构支持

工作流自动构建以下架构：
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM64/Graviton)

Docker 会自动为你的平台拉取正确的架构。

## 在 Databricks 中使用

### 方法 1: 使用 GHCR 镜像

```json
{
  "docker_image": {
    "url": "ghcr.io/<username>/gp-service:latest",
    "basic_auth": {
      "username": "<github-username>",
      "password": "<github-personal-access-token>"
    }
  }
}
```

**创建 GitHub Personal Access Token：**
1. GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate new token (classic)
3. 选择权限：`read:packages`
4. 复制生成的 token
5. 在 Databricks 中配置为 secret

### 方法 2: 使用 Docker Hub 镜像（公开）

```json
{
  "docker_image": {
    "url": "<username>/alpha-detector-service:latest"
  }
}
```

公开镜像无需认证配置。

## 本地测试

### 测试构建流程

```bash
# 安装 act (本地运行 GitHub Actions)
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# 本地运行工作流
act push

# 运行特定 job
act push -j build-and-push
```

### 手动构建多架构镜像

```bash
# 创建 buildx builder
docker buildx create --name mybuilder --use

# 构建并推送
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile.service \
  -t <username>/alpha-detector-service:test \
  --push .
```

## 高级配置

### 自定义构建参数

编辑 `.github/workflows/docker-build.yml`：

```yaml
build-args: |
  VERSION=${{ github.ref_name }}
  COMMIT=${{ github.sha }}
  BUILD_DATE=${{ github.event.head_commit.timestamp }}
  GO_VERSION=1.21
```

### 添加构建通知

在工作流末尾添加：

```yaml
- name: Send notification
  if: always()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### 定时构建

在 `on:` 部分添加：

```yaml
on:
  schedule:
    - cron: '0 2 * * 0'  # 每周日 2:00 AM
```

## 监控和调试

### 查看构建日志

1. 进入仓库的 "Actions" 标签页
2. 选择具体的工作流运行
3. 点击具体的 job 查看详细日志

### 常见问题

**Q: 构建失败，提示权限错误**
A: 检查 workflow 的 permissions 设置，确保包含 `packages: write`

**Q: 推送到 Docker Hub 失败**
A: 验证 `DOCKERHUB_USERNAME` 和 `DOCKERHUB_TOKEN` secrets 配置正确

**Q: 多架构构建很慢**
A: 使用 QEMU 模拟 ARM64 会较慢，考虑使用 GitHub 的 ARM64 runners

**Q: 如何加速构建？**
A: 工作流已启用 GitHub Actions cache，重复构建会自动使用缓存

## 成本和配额

### GitHub Actions 免费额度

**Public 仓库:**
- 无限制的构建时间
- 无限制的存储空间

**Private 仓库:**
- 每月 2000 分钟构建时间
- 500 MB 包存储空间

### GitHub Packages 存储

**免费额度:**
- Public 包：无限制
- Private 包：500 MB

超出后按使用量计费。

## 最佳实践

1. **使用语义化版本标签**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **定期清理旧镜像**
   - 进入 Packages → Package settings
   - 配置自动删除策略

3. **保护 secrets**
   - 仅授予必要的权限
   - 定期轮换 tokens
   - 使用环境 secrets 隔离生产/测试

4. **监控构建状态**
   - 在 README 中添加状态徽章：
     ```markdown
     ![Docker Build](https://github.com/<username>/gp/workflows/Docker%20Build%20and%20Push/badge.svg)
     ```

5. **使用镜像扫描**
   - 添加 Trivy 或 Snyk 扫描步骤
   - 检测安全漏洞

## 完整示例

### 发布新版本

```bash
# 1. 提交代码更改
git add .
git commit -m "Add new feature"

# 2. 创建版本标签
git tag -a v1.2.0 -m "Release v1.2.0"

# 3. 推送到 GitHub
git push origin main --tags

# 4. GitHub Actions 自动构建并推送镜像
# 镜像标签：
# - ghcr.io/<username>/gp-service:latest
# - ghcr.io/<username>/gp-service:v1.2.0
# - ghcr.io/<username>/gp-service:v1.2
# - ghcr.io/<username>/gp-service:v1
```

### 在 Databricks 中使用最新镜像

```python
# 1. 配置 Databricks Job
job_config = {
    "name": "Alpha Detector Service",
    "docker_image": {
        "url": "ghcr.io/<username>/gp-service:latest",
        "basic_auth": {
            "username": "{{secrets/github/username}}",
            "password": "{{secrets/github/token}}"
        }
    },
    "spark_env_vars": {
        "RUN_MODE": "all",
        "ALPHA_KIMI_API_KEY": "{{secrets/alpha-detector/kimi-api-key}}"
    },
    "new_cluster": {
        "node_type_id": "m6g.xlarge",
        "num_workers": 0,
        "spark_version": "13.3.x-scala2.12"
    }
}

# 2. 创建 job
from databricks.sdk import WorkspaceClient
w = WorkspaceClient()
job = w.jobs.create(**job_config)

# 3. 运行 job
run = w.jobs.run_now(job_id=job.job_id)
```

## 参考资源

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Buildx 文档](https://docs.docker.com/buildx/working-with-buildx/)
- [Databricks Container Service](https://docs.databricks.com/en/compute/custom-containers.html)
