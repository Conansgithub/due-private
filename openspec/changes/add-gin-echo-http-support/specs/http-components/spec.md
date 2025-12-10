## ADDED Requirements

### Requirement: 提供 Gin HTTP 组件
Gin 组件 MUST implement `component.Component` lifecycle (Init/Start/Close/Destroy) and SHALL expose Router/Context so handlers mirror the Fiber developer experience.

#### Scenario: 在容器中启动 Gin 服务
- **WHEN** 开发者创建 Gin 组件，设置名称和监听地址，注册至少一个 GET 路由
- **AND** 将组件添加到容器并调用 `Serve()`
- **THEN** HTTP 服务器 MUST 启动成功，路由可访问并返回 200 状态码
- **AND** 组件在收到关闭信号时 SHALL 优雅退出

### Requirement: 提供 Echo HTTP 组件
Echo 组件 MUST implement `component.Component` lifecycle (Init/Start/Close/Destroy) and SHALL expose Router/Context to keep parity with Fiber usage.

#### Scenario: 在容器中启动 Echo 服务
- **WHEN** 开发者创建 Echo 组件，设置名称和监听地址，注册至少一个 GET 路由
- **AND** 将组件添加到容器并调用 `Serve()`
- **THEN** HTTP 服务器 MUST 启动成功，路由可访问并返回 200 状态码
- **AND** 组件在收到关闭信号时 SHALL 优雅退出

### Requirement: 配置项与 Fiber 组件对齐
Gin/Echo 组件 MUST support the same key options as Fiber (监听地址、名称、TLS 证书/密钥、日志开关、超时/限流钩子、中间件注入接口) to allow drop-in replacement.

#### Scenario: 应用与 Fiber 等价的配置
- **WHEN** 开发者为 Gin/Echo 组件配置 TLS 证书、开启请求日志，并注册自定义中间件
- **THEN** 服务 MUST 以 HTTPS 启动，日志 SHALL 按配置输出，且自定义中间件被调用

### Requirement: 支持常用中间件
Gin/Echo 组件 MUST provide built-in or pluggable support for CORS、Recovery、Logger、请求ID 等常用中间件, with behavior aligned to Fiber or documented differences.

#### Scenario: 启用 CORS 与请求日志
- **WHEN** 开发者通过组件选项开启 CORS 与 Logger
- **THEN** 跨域请求 MUST 获得预期的 CORS 头，访问日志 SHALL 按配置输出

### Requirement: 文档与示例
Documentation MUST include minimal runnable examples for Gin and Echo, common middleware samples, and SHALL note differences versus Fiber and migration tips.

#### Scenario: 开发者参照示例快速启动
- **WHEN** 开发者复制文档中的最小示例代码并执行
- **THEN** Gin 或 Echo 服务器 MUST 成功启动并响应示例路由
- **AND** 文档 SHALL 列出与 Fiber 的主要差异与迁移提示
