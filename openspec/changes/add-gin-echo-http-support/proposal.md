# 接入 Gin 与 Echo HTTP 组件

## 变更ID
add-gin-echo-http-support

## 为什么
Fiber 已作为默认 HTTP 组件，但部分团队更熟悉 Gin 或 Echo，且希望复用其成熟的中间件生态。为了降低迁移成本并提供框架选择，需要为 due 增加与 Fiber 等价的 Gin / Echo 组件。

## 什么变更
- 新增 `component/http/gin` 与 `component/http/echo` 组件，实现 `component.Component` 接口，生命周期与 Fiber 一致。
- 提供与 Fiber 组件对齐的配置项（地址、名称、TLS、日志、中间件、超时等）。
- 暴露路由与上下文封装，保持与 Fiber 版的开发体验一致（Router() / Context）。
- 支持容器统一管理：Init/Start/Close/Destroy 以及优雅退出。
- 提供示例与文档，展示 Gin/Echo 的最小可运行用法和中间件用法。

## 影响
- 受影响代码：`component/http/*`、容器注册/工厂相关代码、示例与文档。
- 受影响规范：`specs/http-components` 将增加 Gin / Echo 组件的要求与场景。

## 成功标准
1) 通过容器即可启动 Gin/Echo HTTP 服务器，完成 Init/Start/Close/Destroy 全生命周期。
2) 配置项与 Fiber 组件保持同级覆盖（含 TLS、日志、中间件、超时/限流 Hooks）。
3) Router/Context 接口与 Fiber 组件形态一致，方便迁移示例代码。
4) 文档含最小示例、常用中间件示例（CORS/Recovery/Logger），并说明差异点。
5) `openspec validate --strict` 通过；新增组件具备基础集成测试。

## 风险与缓解
- **接口偏差**：Gin/Echo 配置或上下文封装与 Fiber 不一致 → 设计期对齐接口，新增集成测试覆盖。
- **中间件兼容性**：生态差异导致行为不同 → 文档列出差异与默认中间件映射，提供可选钩子。
- **性能差异**：Gin/Echo 性能低于 Fiber → 在文档中注明适用场景并可添加基准链接。
