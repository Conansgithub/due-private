# due Fiber（HTTP）组件使用指南（超详细版）

本指南讲清楚 **due 框架下的 Fiber HTTP 组件如何使用**、**如何从零创建一个 Web 站点**，以及常用高级能力与配置。文档以 due 开源主线为准（Fiber 组件是默认 Web 组件）。

> 说明：本仓库为 `due-private`，对应开源仓库 `due` 的同名模块。示例中默认使用开源 import 路径；如你在私有仓库使用，请把 `github.com/Conansgithub/due/...` 替换成 `github.com/Conansgithub/due-private/...`。

---

## 1. 组件定位与关键概念

### 1.1 due 的组件化容器

due 用 **Container 容器**统一管理组件生命周期：

1. `Init()`：初始化（读配置、参数检查、准备资源）
2. `Start()`：启动（监听端口、注册服务、开启 goroutine）
3. 等待系统信号（默认会阻塞）
4. `Close()`：关闭（容器收到退出信号后并发调用）
5. `Destroy()`：销毁（Close 后做最后清理）

容器实现见 `container.go`：`NewContainer()` → `Add()` → `Serve()` 会按上述顺序执行所有组件。

### 1.2 Fiber HTTP 组件是什么

due 的 Web 组件位于：

- 开源：`github.com/Conansgithub/due/component/http/v2`
- 私有镜像：`github.com/Conansgithub/due-private/component/http/v2`

该组件基于 **Fiber v3**，封装成 due 组件，提供：

- 与容器对齐的生命周期（当前版本主要实现了 Start）
- Options 配置读取（`etc.http.*`）
- `Proxy()` 暴露统一开发接口
- `Router()` 与 `Context` 适配，让路由与业务代码不直接绑定 Fiber
- 内置常用中间件（Logger / Recovery / CORS / Swagger）

### 1.3 Proxy / Router / Context

创建 HTTP Server 后，你主要通过 `Proxy()` 使用：

- `Proxy.App()`：拿到底层 *fiber.App（原生 API）
- `Proxy.Router()`：拿到 due Router（对齐 due Handler 体验）
- `Proxy.NewMeshClient()`：创建微服务客户端（需注入 transporter）

`Router()` 提供与 Fiber 组件一致的路由表面：`Get/Post/.../Group/Add/All`。

`Context` 继承 `fiber.Ctx` 并新增：

- `CTX()`：返回原生 `fiber.Ctx`
- `Proxy()`：返回当前 HTTP Proxy
- `Success()` / `Failure()`：统一 JSON 响应（含 code/message）
- `StdRequest()`：转成 `*net/http.Request`

---

## 2. 安装与版本对齐

### 2.1 安装 due 主模块 + http 组件

```bash
go get github.com/Conansgithub/due/v2@latest
go get github.com/Conansgithub/due/component/http/v2@latest
```

### 2.2 版本一致性（非常重要）

due 是模块化仓库，各子模块单独发版。**主模块与子模块版本需保持一致**，否则可能出现 API 不兼容或运行期异常。

做法：

1. 先确定你要用的 due 主版本（例如 `v2.3.2`）。
2. 去 release 页面查对应 commit。
3. 用该 commit 号拉取子模块：

```bash
go get github.com/Conansgithub/due/component/http/v2@<commit>
```

详细原因与排查方法见 due README「版本不一致问题」章节。

---

## 3. 配置系统（etc）与 HTTP 配置项

### 3.1 etc 是什么

`etc` 用来承载 **项目启动配置**，仅用于静态启动参数（不可热改）：

- 默认读取 `./etc` 目录
- 支持 `toml/yaml/json/xml` 等格式
- 可通过环境变量 `DUE_ETC` 或启动参数 `--etc` 改路径

源码见 `etc/etc.go`。

### 3.2 HTTP 组件读取哪些配置

Fiber HTTP 组件在初始化时会读取以下键（默认前缀 `etc.http.*`）：

| 配置键 | 说明 | 默认值 |
|---|---|---|
| `etc.http.name` | 服务名称（也会写入 ServerHeader） | `"http"` |
| `etc.http.addr` | 监听地址 | `":8080"` |
| `etc.http.console` | 是否打印请求日志（Logger） | `false` |
| `etc.http.bodyLimit` | 请求体最大大小（支持单位） | `4M` |
| `etc.http.concurrency` | 最大并发连接数（当前版本未显式传入 Fiber Config，仅保留配置位） | `262144` |
| `etc.http.strictRouting` | 是否严格路由（`/foo` 与 `/foo/` 区分） | `false` |
| `etc.http.caseSensitive` | 是否大小写敏感 | `false` |
| `etc.http.certFile` | TLS 证书文件路径 | `""` |
| `etc.http.keyFile` | TLS 私钥文件路径 | `""` |
| `etc.http.cors.*` | CORS 跨域配置 | 见下 |
| `etc.http.swagger.*` | Swagger 文档配置 | 见下 |

配置结构体位于 `component/http/options.go`，默认值可直接对照源码。

### 3.3 推荐的 etc/http.toml 模板

在 `./etc/etc.toml` 里加入：

```toml
[http]
  name = "http"
  addr = ":8080"
  console = true
  bodyLimit = "4M"
  concurrency = 262144
  strictRouting = false
  caseSensitive = false
  keyFile = ""
  certFile = ""

  [http.cors]
    enable = false
    allowOrigins = []
    allowMethods = []
    allowHeaders = []
    allowCredentials = false
    exposeHeaders = []
    maxAge = 0
    allowPrivateNetwork = false

  [http.swagger]
    enable = false
    title = "API文档"
    basePath = "/swagger"
    filePath = "./docs/swagger.json"
    swaggerBundleUrl = ""
    swaggerPresetUrl = ""
    swaggerStylesUrl = ""
```

> 说明：HTTP 组件只会读取 `etc.http.*`；如果你用多环境，可以用不同配置文件覆盖同名字段。

---

## 4. 从零创建一个 due Web 站点（Hello World）

### 4.1 目录结构建议

```
my-web/
├─ go.mod
├─ main.go
├─ etc/
│  └─ etc.toml
└─ docs/
   └─ swagger.json   # 可选（启用 swagger 时必须存在）
```

### 4.2 main.go 最小可运行示例

```go
package main

import (
    due "github.com/Conansgithub/due/v2"
    http "github.com/Conansgithub/due/component/http/v2"
)

func main() {
    // 1) 创建容器
    container := due.NewContainer()

    // 2) 创建 Fiber HTTP 组件
    server := http.NewServer(
        http.WithName("web"),
        http.WithAddr(":8080"),
        http.WithConsole(true),
    )

    // 3) 注册路由（推荐用 due Router）
    router := server.Proxy().Router()
    router.Get("/health", func(ctx http.Context) error {
        return ctx.Success(map[string]any{"ok": true})
    })

    // 4) 把组件加入容器
    container.Add(server)

    // 5) 启动容器（阻塞直到收到退出信号）
    container.Serve()
}
```

运行：

```bash
go run main.go
```

访问 `http://127.0.0.1:8080/health` 会返回：

```json
{"code":200,"message":"OK","data":{"ok":true}}
```

### 4.3 每一步发生了什么

1. `NewContainer()`：创建一个空容器。
2. `http.NewServer()`：构建 Fiber App 并加载 options：
   - 读取 `etc.http.*` 的默认值
   - 叠加你传入的 Option
   - 初始化 Fiber App，自动挂载内置中间件
3. `Proxy().Router()`：拿到 due Router，后续路由全部注册到 Fiber App。
4. `container.Add(server)`：登记组件。
5. `container.Serve()`：
   - 保存 pid、打印框架信息
   - 调用 `server.Init()`（HTTP 组件当前为空）
   - 调用 `server.Start()` 启动监听（goroutine）
   - 阻塞等待 SIGINT/SIGTERM
   - 收到信号后调用 Close/Destroy（HTTP 组件当前为空实现）

---

## 5. 路由注册：due Handler 与 Fiber 原生 Handler

### 5.1 两类 Handler 你都可以用

Router 方法的 `handler any` 参数接受：

1. **due Handler**：`func(ctx http.Context) error`
2. **fiber.Handler**：`func(c fiber.Ctx) error`

示例：

```go
router.Get("/due", func(ctx http.Context) error {
    return ctx.Success("hello due handler")
})

router.Get("/fiber", func(c fiber.Ctx) error {
    return c.SendString("hello fiber handler")
})
```

### 5.2 常用 Router API

Router 接口（与 Fiber 一致）：

- `Get/Post/Put/Delete/Head/Patch/Options/Connect/Trace`
- `All(path, handler, middlewares...)`
- `Add(methods []string, path string, handler any, middlewares ...any)`
- `Group(prefix, middlewares...)`

### 5.3 路由组 Group

```go
api := router.Group("/api")

api.Get("/users/:id", func(ctx http.Context) error {
    id := ctx.Params("id")
    return ctx.Success(map[string]string{"id": id})
})
```

你也可以为 Group 传中间件：

```go
api := router.Group("/api", myAuthMiddleware)
```

中间件类型规则与 Handler 完全一致（见第 7 节）。

### 5.4 复合方法 Add

```go
router.Add([]string{"GET","POST"}, "/ping", func(ctx http.Context) error {
    return ctx.Success("pong")
})
```

---

## 6. Context 详细能力

### 6.1 读取请求

因为 `http.Context` 继承 Fiber 的 `fiber.Ctx`，所以 Fiber 的所有读写 API 都可以直接用：

```go
router.Post("/login", func(ctx http.Context) error {
    body := ctx.Body()
    ua := ctx.Get("User-Agent")
    q := ctx.Query("from")
    return ctx.Success(map[string]any{
        "body": string(body),
        "ua": ua,
        "from": q,
    })
})
```

### 6.2 Success / Failure

`Success` 与 `Failure` 用来统一 JSON 响应体格式：

```go
router.Get("/ok", func(ctx http.Context) error {
    return ctx.Success(map[string]any{"x": 1})
})

router.Get("/bad", func(ctx http.Context) error {
    return ctx.Failure(errors.New("something wrong"))
})
```

`Failure(rst any)` 支持三种输入：

- `error`：会被 `codes.Convert(err)` 映射为 due 业务码
- `*codes.Code`：直接使用该 code/message
- 其他类型：返回 `codes.Unknown`

> 注意：Fiber 版本的 `Success/Failure` **只影响 JSON body 中的 code/message**，HTTP Status 默认仍然是 200（除非你自己设置）。

### 6.3 StdRequest

某些第三方库只接受 `*http.Request` 时可用：

```go
router.Get("/std", func(ctx http.Context) error {
    stdReq := ctx.StdRequest()
    _ = stdReq // 传给你自己的 net/http 逻辑
    return ctx.Success("ok")
})
```

### 6.4 CTX / Proxy

如果你要访问原生 Fiber 细节或 Proxy：

```go
router.Get("/raw", func(ctx http.Context) error {
    fiberCtx := ctx.CTX()
    _ = fiberCtx

    proxy := ctx.Proxy()
    _ = proxy

    return ctx.Success("ok")
})
```

---

## 7. 中间件（全局 / 分组 / 单路由）

### 7.1 内置中间件

HTTP 组件内部会自动挂载：

1. **Recovery（必开）**
   - Fiber `recover.New(recover.Config{EnableStackTrace:true})`
   - 捕获 panic 并输出栈

2. **Logger（可选）**
   - 由 `console` 配置控制
   - `etc.http.console=true` 或 `WithConsole(true)`

3. **CORS（可选）**
   - 由 `etc.http.cors.enable` 控制
   - 字段见 3.3 配置模板

4. **Swagger（可选）**
   - 由 `etc.http.swagger.enable` 控制
   - 启用时会挂载 swagger 中间件（见第 8 节）

### 7.2 全局中间件注入（WithMiddlewares）

创建 Server 时可以注入自定义全局中间件：

```go
server := http.NewServer(
    http.WithMiddlewares(
        // 1) due Handler 形式
        func(ctx http.Context) error {
            ctx.Set("X-App", "demo")
            return ctx.Next()
        },
        // 2) fiber.Handler 形式
        func(c fiber.Ctx) error {
            c.Set("X-Fiber", "yes")
            return c.Next()
        },
    ),
)
```

中间件按传入顺序执行。

### 7.3 分组/单路由中间件

Router 的每个方法都接受 `middlewares ...any`：

```go
router.Get(
    "/private",
    privateHandler,
    authMiddleware,
    rateLimitMiddleware,
)
```

这些中间件类型与全局一致：支持 due Handler 或 fiber.Handler。

---

## 8. Swagger 文档

### 8.1 开启 Swagger

1. 准备 swagger spec 文件（json 或 yaml），例如 `./docs/swagger.json`
2. 在 etc 中开启：

```toml
[http.swagger]
  enable = true
  title = "API 文档"
  basePath = "/swagger"
  filePath = "./docs/swagger.json"
```

启动后访问：

- UI：`http://<host>:<port>/swagger`
- Spec：`http://<host>:<port>/swagger/docs/swagger.json`（内部会把 basePath + filePath 拼接）

> 注意：`filePath` 必须存在，否则中间件会 `Fatalf` 直接退出进程。

### 8.2 CDN 资源替换

当默认 CDN 不可用时可替换：

```toml
[http.swagger]
  enable = true
  swaggerBundleUrl = "https://your.cdn/swagger-ui.js"
  swaggerPresetUrl = "https://your.cdn/swagger-ui-standalone-preset.js"
  swaggerStylesUrl = "https://your.cdn/swagger-ui.css"
```

---

## 9. HTTPS / TLS

### 9.1 通过配置启用

```toml
[http]
  addr = ":8443"
  certFile = "./certs/server.crt"
  keyFile = "./certs/server.key"
```

满足 `certFile` 与 `keyFile` 都非空时，Fiber 会自动以 HTTPS 监听。

### 9.2 通过 Option 启用

```go
server := http.NewServer(
    http.WithAddr(":8443"),
    http.WithCredentials("./certs/server.crt", "./certs/server.key"),
)
```

---

## 10. 启动信息与监听地址解析

HTTP 组件在 `Start()` 里会：

1. 调用 `xnet.ParseAddr(opts.addr)`
2. 得到：
   - `listenAddr`：真正监听地址
   - `exposeAddr`：用于打印/注册的暴露地址

规则：

- `addr=":8080"` 或 `addr="0.0.0.0:8080"`：监听 `0.0.0.0:8080`，打印/暴露为 **私网 IP:8080**
- `addr="127.0.0.1:8080"`：监听并暴露该固定地址
- `addr=":0"` 或 `addr="0.0.0.0:0"`：随机端口，并打印分配后的端口

---

## 11. 与 Registry / Transporter 协作（可选）

HTTP 组件支持注入：

- `registry.Registry`：服务注册/发现
- `transport.Transporter`：RPC 客户端/服务端构建器

当两者同时存在时，`Start()` 会执行：

```go
transporter.SetDefaultDiscovery(registry)
```

这样你就可以用服务发现模式创建 mesh client：

```go
server := http.NewServer(
    http.WithRegistry(consul.NewRegistry()),
    http.WithTransporter(grpc.NewTransporter()),
)

router.Get("/call-mesh", func(ctx http.Context) error {
    client, err := ctx.Proxy().NewMeshClient("discovery://user-service")
    if err != nil {
        return ctx.Failure(err)
    }
    // client.Call(...) 依据你选用的 transporter 协议
    return ctx.Success("ok")
})
```

---

## 12. 使用 Fiber 原生能力（高级）

当 due Router 不够用时，可以直接操作原生 App：

```go
app := server.Proxy().App()

app.Static("/assets", "./public")

app.Get("/native", func(c fiber.Ctx) error {
    return c.SendString("native fiber route")
})
```

原生能力示例（按 Fiber 官方文档扩展）：

- 静态文件 / 模板渲染
- WebSocket
- 自定义错误处理器
- 自定义 BodyParser / Validator
- 自建子路由器

由于 `App()` 直接暴露 *fiber.App，你可以无损使用 Fiber 生态。

---

## 13. 生命周期与优雅退出（当前版本注意点）

截至当前版本，Fiber HTTP 组件只实现了 `Start()`，`Close()`/`Destroy()` 继承自 `component.Base` 为 **空实现**。

含义：

- 容器退出时不会主动调用 Fiber 的 Shutdown
- 进程退出会直接终止监听

如果你的业务需要优雅关闭（如等待在途请求），可以：

1. 自己封装一层组件并覆盖 Close：

```go
type Web struct{ *http.Server }

func (w *Web) Close() {
    // 调用 Fiber 的优雅关闭 API（具体方法以 Fiber v3 文档为准）
    // _ = w.Proxy().App().ShutdownWithTimeout(...)
}
```

2. 或者自己监听系统信号，在 `container.Serve()` 返回前手动关停。

---

## 14. 常见问题与排错

1. **服务启动报错：swagger 文件不存在**
   - 检查 `etc.http.swagger.filePath`
   - 文件必须存在，否则 swagger 中间件会直接 `Fatalf`

2. **路由中间件不生效**
   - 只支持 `fiber.Handler` 或 `http.Handler`
   - 传入其他类型会被忽略

3. **HTTP code 和业务 code 混淆**
   - `Success/Failure` 只写 JSON body 内的 `code/message`
   - 若要改变 HTTP Status，请自行 `ctx.Status(x)` 或使用 Fiber 原生接口

4. **主模块/子模块版本不一致**
   - 现象：编译不过或运行期异常
   - 解决：按第 2.2 节对齐 commit

---

## 15. 附录：完整最小示例（可直接复制）

`main.go`：

```go
package main

import (
    "github.com/gofiber/fiber/v3"

    due "github.com/Conansgithub/due/v2"
    http "github.com/Conansgithub/due/component/http/v2"
)

func main() {
    container := due.NewContainer()

    server := http.NewServer(
        http.WithName("demo-web"),
        http.WithAddr(":8080"),
        http.WithConsole(true),
    )

    r := server.Proxy().Router()

    r.Get("/", func(ctx http.Context) error {
        return ctx.Success("hello due + fiber")
    })

    r.Get("/native", func(c fiber.Ctx) error {
        return c.SendString("hello native fiber")
    })

    api := r.Group("/api")
    api.Get("/ping", func(ctx http.Context) error {
        return ctx.Success("pong")
    })

    container.Add(server)
    container.Serve()
}
```

`etc/etc.toml`（最小）：

```toml
[http]
  addr = ":8080"
  console = true
```

---

如果你希望我再补一份 “Echo/Gin 版同结构文档” 或把这份拆成 OpenSpec 的 Requirement/Scenario 规范形式，也可以继续说。  
