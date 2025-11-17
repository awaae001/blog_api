# Blog API 管理面板

基于 Vue 3 + TypeScript + Element Plus 的管理后台。

## 技术栈

- **Vue 3** - 前端框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Element Plus** - UI 组件库
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Axios** - HTTP 客户端

## 快速开始

### 安装依赖

```bash
cd web
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:5173/panel/login

### 生产构建

```bash
npm run build
```

构建输出会自动放置到 `../data/panel` 目录，后端会自动提供静态文件服务。

## 项目结构

```
web/
├── src/
│   ├── api/           # API 接口定义
│   │   └── auth.ts    # 认证接口
│   ├── components/    # 可复用组件
│   ├── router/        # 路由配置
│   │   └── index.ts   # 路由定义
│   ├── utils/         # 工具函数
│   │   └── request.ts # Axios 封装
│   ├── views/         # 页面组件
│   │   ├── Login.vue  # 登录页面
│   │   └── Panel.vue  # 管理面板主页
│   ├── App.vue        # 根组件
│   └── main.ts        # 入口文件
├── index.html
├── vite.config.ts     # Vite 配置
├── tsconfig.json      # TypeScript 配置
└── package.json
```

## 路由说明

- `/panel/login` - 登录页面
- `/panel` - 管理面板主页（需要认证）

## 认证流程

1. 用户在登录页面输入用户名和密码
2. 前端调用 `/api/verify` 接口验证
3. 后端验证成功后返回 JWT token
4. 前端将 token 存储在 localStorage
5. 后续请求自动在 Authorization header 中携带 token
6. 如果 token 过期或无效，自动跳转到登录页

## 开发说明

### 配置代理

开发模式下，Vite 会将 `/api` 开头的请求代理到后端服务器（http://localhost:10024）。

### 添加新页面

1. 在 `src/views/` 创建新的 Vue 组件
2. 在 `src/router/index.ts` 添加路由配置
3. 在 `Panel.vue` 的侧边栏菜单中添加导航项

### 添加新接口

1. 在 `src/api/` 目录下创建对应的 API 文件
2. 定义 TypeScript 接口类型
3. 使用封装好的 `request` 实例发起请求

## 环境配置

后端需要在 `.env` 文件中配置：

```env
WEB_PANEL_USER = "admin"     # 管理员用户名
WEB_PANEL_PWD = "password"   # 管理员密码
JWT_SECRET = ""               # JWT 密钥（可选，留空自动生成）
```

CORS 配置需要在 `data/config/system_config.json` 中添加允许的域名：

```json
{
  "system_conf": {
    "safe_conf": {
      "cors_allow_hostlist": [
        "http://localhost:5173"
      ]
    }
  }
}
```
