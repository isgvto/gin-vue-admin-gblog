# blog-view

博客展示前台，保留 Vue 2 技术栈。完整的数据库初始化、后端与后台启动说明见 [根目录 README](../README.md)。

## 安装与运行

环境要求：Node.js 24+、npm 10+。

```sh
npm ci
npm run serve -- --port 8081
```

开发请求通过 `/api` 代理到 `http://127.0.0.1:8888`，代理配置位于 `vue.config.js`。

## 构建与回归

```sh
npm run build
```

输出到 `dist/`。浏览器回归直接加载该生产产物，模拟所有业务 API，不连接真实后端、不提交真实评论。需要 Python、Playwright 和本机 Chrome：

```sh
python -m pip install playwright
npm run test:browser
# 使用 Edge：
npm run test:browser -- --channel msedge
```

测试覆盖：文章与分类/标签路由、日期、接口参数与身份标识、评论校验与 HTML 转义、密码文章、搜索、图片查看与关闭、Mermaid、文档页、动态点赞及浏览器运行时错误。

## 依赖升级说明

本轮更新直接依赖，并通过 `npm update` 更新版本约束允许的传递依赖，保留 Vue 2、Vue Router 3、Vuex 3、Element UI 2 和 Mermaid 11。直接升级的版本固定在 `package.json`，安装结果固定在 `package-lock.json`；部署使用 `npm ci`。

| 依赖 | 升级前 | 升级后 |
| --- | --- | --- |
| Vue / vue-template-compiler | 2.6.11 | 2.7.16 |
| Vue Router | 3.3.4 | 3.6.5 |
| Vuex | 3.5.1 | 3.6.2 |
| Element UI | 2.13.2 | 2.15.14 |
| Axios | 0.24.0 | 1.20.0 |
| core-js（直接依赖） | 3.6.5 | 3.50.0 |
| sanitize-html | 2.3.3 | 2.18.0 |
| Moment | 2.27.0 | 2.31.0 |
| Semantic UI CSS | 2.4.1 | 2.5.0 |
| v-viewer（Vue 2 分支） | 1.5.1 | 1.7.4 |
| Vue CLI 服务与三个插件 | 5.0.8 | 5.0.9 |

Vue 2 的 `vue-loader@15` 与更新的 webpack 组合会产生样式默认导出告警，因此通过 `overrides` 将 webpack 固定为已验证的 `5.106.2`，其余传递依赖仍按兼容版本更新。后续升级 webpack 时需要复验该组合，不直接修改组件样式或隐藏告警。

Vue 2 已结束官方维护，Vue CLI 处于维护模式。安装时仍可能出现 `stable`、`inflight`、`glob`、`rimraf`、`consolidate`、旧 Babel 插件和传递依赖 `core-js@2` 的弃用告警，分别来自 Vue CLI 构建链和 Element UI。不能仅为消除告警而强制覆盖这些包的大版本。

2026-10-03 使用官方 npm registry 审计，报告从 50 项降至 27 项，critical 从 2 项降至 0 项；剩余为 low 4、moderate 9、high 14，包含开发依赖及 Vue 2 生态依赖。这是当日审计结果，不代表零漏洞或全部可在当前技术栈内修复。后续复查：

```sh
npm audit --registry=https://registry.npmjs.org
```

不使用 `npm audit fix --force` 自动处理剩余问题；先核对受影响依赖、兼容性及实际使用方式，再决定升级或迁移。

参考：[Vue 2.7 升级说明](https://v2.vuejs.org/v2/guide/migration-vue-2-7.html)、[Vue CLI 文档](https://cli.vuejs.org/)。
