## 1. 实施
- [ ] 1.1 设计 Gin/Echo 组件接口，确保与 Fiber 版配置/生命周期对齐
- [ ] 1.2 实现 `component/http/gin` 组件：Server、Router、Context 适配、配置项与选项函数
- [ ] 1.3 实现 `component/http/echo` 组件：Server、Router、Context 适配、配置项与选项函数
- [ ] 1.4 容器集成：注册组件、支持 Init/Start/Close/Destroy、优雅退出
- [ ] 1.5 中间件与常用功能：CORS、Recovery、Logger、请求ID、限流/超时钩子（与 Fiber 对齐）
- [ ] 1.6 示例与文档：最小可运行示例、常用中间件示例、差异说明
- [ ] 1.7 测试：为 Gin/Echo 组件添加集成测试（启动、路由、优雅关闭、配置生效）
- [ ] 1.8 运行 `openspec validate --strict` 并修复问题
