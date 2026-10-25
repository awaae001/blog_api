# blog_api

是时候给 memos / fcircle / 随机图 api 扔到 土立土及木甬 里面去了

一个基于 `Go + Gin + SQLite + Vue3` 的博客周边 API 服务，覆盖：`memos`（动态）、`fcircle`（友链与 RSS 聚合）、`随机图 API`（图片仓库与图片直链）。

> 我有我的便利，你有你的自由
> 
> 后端 API 可独立运行，不依赖内置 `web` 面板；你也可以完全自己做一个面板对接这些接口。

## 这项目到底做了啥

- `memos`：动态内容（moments）增删改查、媒体绑定、点赞/取消点赞。
- `fcircle`：友链管理、友链 RSS 拉取、RSS 文章列表对外输出。
- `随机图 API`：图片入库、更新、删除、公开访问（`/api/public/image/*id`）。
- 资源管理：支持本地资源上传/删除，也支持 OSS 上传/删除。
- 管理后台：内置 `web/`（Vue3 + Element Plus）仅作为默认实现，可不用或自行替换。
- 反机器人能力：支持 Turnstile、指纹签发、邮箱验证码（按配置启用）。

## 快速启动

- 复制 `.env.example` 到 `.env` 并按需修改。
- 首次启动会自动在 `data/config/` 生成 `system_config.json` 和 `friend_list.json`，不会覆盖已有文件。
- 按需编辑 `friend_list.json`，用于空数据库的友链初始化。

需要 Go 1.26+ 和 C 编译器（SQLite / WebP 使用 CGO）。

```bash
go run main.go
```

默认监听：`0.0.0.0:10024`。

> 通过管理面板保存公开 API 功能开关后会立即热生效。监听地址、数据库路径、机器人集成、OSS 等启动型配置仍可能需要重启；保存接口会返回 `restart_required_keys` 提示。

然后启动可选的管理面板

```bash
cd web
pnpm install --frozen-lockfile
pnpm dev
```

开发访问：`http://localhost:5173/panel/login`。

构建发布：

```bash
cd web
pnpm build
```

构建产物输出到 `data/panel`，由后端统一托管。

## 公开 API 功能开关

公开业务 API 默认全部关闭。可在管理面板的“系统设置 → 功能开关”中选择启用，或编辑 `data/config/system_config.json`：

```json
{
  "system_conf": {
    "safe_conf": {
      "enabled_public_apis": ["moments", "image", "friend", "rss", "email"]
    }
  }
}
```

支持以下 key：

- `moments`：动态读取及 reaction 接口。
- `image`：随机图和指定图片公开接口。
- `friend`：公开友链读取及邮箱 token 自助管理接口。
- `rss`：公开 RSS 文章读取接口。
- `email`：发送/验证邮箱验证码及签发邮箱 token 的接口。

未启用的接口统一返回 HTTP 404。未知 key（包括错误拼写）会被忽略；管理接口、管理员登录、Turnstile、指纹和公开验证配置不受此列表影响。通过管理面板保存该列表可立即生效，无需重启。