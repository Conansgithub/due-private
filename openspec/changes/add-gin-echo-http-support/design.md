# 设计: Gin / Echo HTTP 组件接入

## 上下文
- 现有 Fiber 组件位于 `component/http/`，使用 options + proxy + router + context 的固定模式，生命周期遵循 `component.Component`（Base+Init/Start/Close/Destroy）。
- 配置默认读取 `etc.http.*`，包含 addr/name/tls/log/cors/swagger 等。
- Router/Context 将业务 handler 适配为统一接口，Proxy 暴露 Transporter/Router/App。

## 目标
- 提供 `component/http/gin` 与 `component/http/echo` 两个组件，接口与 Fiber 版对齐，可直接被容器管理。
- 复用/对齐配置项、生命周期、响应封装、Proxy 能力，降低迁移成本。
- 内置常用中间件（Recovery/Logger/CORS）选项，允许注入自定义中间件。
- 提供最小示例与中间件示例，文档列出与 Fiber 的差异。

## 非目标
- 不在本次实现性能基准；只预留钩子或在文档说明需另行测试。
- 不引入全新的跨组件抽象层；优先直接复用 Fiber 的模式，必要时轻量抽取共用代码。

## 方案概述
- **包结构**
  - `component/http/gin`：`server.go`、`options.go`、`context.go`、`router.go`、`proxy.go`、`status.go`(复用 http/status?)、`middleware.go`(可选)。
  - `component/http/echo`：同上。
  - 如有小共用段（Resp/Status/option keys），可提取到 `component/http/common` 或直接复用现有 `component/http` 类型以免重复。

- **生命周期**
  - 组件结构体持有：options、engine实例(*gin.Engine / *echo.Echo)、proxy。
  - 实现 `Name/Init/Start/Close/Destroy`：
    - Start：解析 addr，打印 info，与 Fiber 类似；支持 TLS（cert/key）；在 goroutine 启动；将 registry 设置到 transporter。
    - Close：调用 `Shutdown(ctx)` 或 `Close()`（Echo 提供 Shutdown）。Destroy 空实现。

- **Options 与配置**
  - 采用函数式选项；字段与 Fiber 对齐：name/addr/console/bodyLimit/concurrency/strictRouting/caseSensitive/certFile/keyFile/registry/transporter/corsOpts/swagOpts/middlewares。
  - 配置键：沿用 `etc.http.*` 会与 Fiber 共享；为避免冲突可增加子前缀，例如 `etc.http.gin.*`、`etc.http.echo.*`；设计上建议：
    - 默认从通用键读取以保持行为一致；若存在框架专属键则覆盖（优先级：框架专属 > 通用 > 默认）。
  - CORS/Swagger 结构体保持字段名一致，便于共享配置结构。

- **Router/Context 适配**
  - Router 接口同 Fiber 版，内部将 Handler 适配到 gin.HandlerFunc / echo.HandlerFunc。
  - Context：包装 gin.Context / echo.Context，提供 `CTX()`、`Proxy()`、`Failure()`、`Success()`、`StdRequest()`；Resp 结构与 Fiber 版一致。
  - Proxy：暴露 Router() / App() / NewMeshClient(target)。

- **中间件策略**
  - 默认启用 Recovery；console/log 开关映射到 gin-contrib/logger 与 echo/middleware.Logger。
  - CORS：使用 gin-contrib/cors 与 echo/middleware.CORS，通过 CorsOptions 映射。
  - 自定义 middlewares：接受框架原生 handler 或统一 Handler（与 Fiber 版一致的签名），在 Add/Use 时包裹。

- **Swagger（可选最小实现）**
  - 若现有 swagger 中间件不可直接复用，先提供禁用默认值，接口保留；实现可以延后或用简单 handler 提供静态文件/URL 代理（与 Fiber 配置项兼容）。

- **容器集成**
  - 与 Fiber 相同：`Proxy()` 返回 proxy，`Start()` 内部处理地址解析与打印，支持 transporter+registry 注入。
  - 优雅关闭：在 Close 内调用 `Shutdown` 并结合 context.WithTimeout；容器发出的 stop 信号时可生效。

- **测试策略（实现时参考）**
  - 启动 Gin/Echo 服务器在随机端口，注册 GET 路由，发起 HTTP 请求验证 200 与 JSON 响应。
  - 验证 CORS 头是否生效（OPTIONS/GET）。
  - 验证 Close 后端口释放。

## 迁移/示例指引
- 最小示例：NewServer + WithName/WithAddr + Router.GET + container.Add + container.Serve。
- 中间件示例：启用 CORS + Logger；演示自定义 middleware（统一 Handler 签名）。
- 记录与 Fiber 差异：
  - Gin/Echo 的中间件链与 Fiber 顺序/错误处理差异。
  - BodyLimit/Concurrency 实现方式可能不同（Echo 有 MaxRequestBodySize；Gin 需手动中间件）。

## 开放问题
- Swagger：使用哪种具体实现？（可先占位 no-op，或引入对应生态包）
- 配置键冲突策略：是否需要独立命名空间（`etc.http.gin.*`），或允许共享？需与使用方沟通。
- 性能选项（例如响应压缩、Prefork/HTTP3）是否需要对齐？当前倾向保持最小集。
