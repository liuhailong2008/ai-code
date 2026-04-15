# OpenCode全局规则

# 目录
- front-demo/style 下文件用于样式参考，页面需同时支持三种样式，并可以随时点击页面按钮切换。
- 前端代码生成到 front-demo/code 目录下。

## 语言和风格

- 始终使用简体中文回复
- 直接回答问题，不要客套话
- 代码注释也用中文

## 工作态度

- 严谨的工作态度，保证完美的质量标准
- 始终用简短的列表告诉我当前在做什么
- 修改代码前先阅读相关文件
- 每次只做最小必要的修改

## 求真原则

- 不确定/信息不足时先查证或提问澄清
- 对环境/配置/源码/行为的结论必须有证据
- 回答里把"事实"和"推测/假设"分开写

## 工作习惯

- 工具安装优先使用Homebrew (路径: /opt/homebrew/bin)
- Homebrew 安装后需配置 PATH: export PATH="/opt/homebrew/bin:$PATH"
- 编写新代码前，先确认项目中是否已有类似实现
- 优先复用现有组件和工具函数，而非新建
- 启动预览服务器后，必须先用 curl 验证返回 200 后再告知用户访问地址

## 使用技能

- 涉及创造性工作时（创建功能、构建组件、添加功能），必须先使用brainstorming技能
- 遇到任何bug或意外行为，必须先使用systematic-debugging技能
- 在实现任何功能或修复bug之前，必须先使用test-driven-development技能
- 需要代码审查时，必须使用requesting-code-review技能
- 收到代码审查反馈时，必须使用receiving-code-review技能

## 前端开发

### frontend-ui-ux Skill 使用方法

在 OpenCode 中进行前端 UI/UX 工作时，使用 skill 工具加载：

```
skill name: frontend-ui-ux
```

该 skill 提供：
- 无需设计稿即可创建专业 UI
- 现代设计趋势和最佳实践
- 专业的配色方案、间距、动画指导
- 响应式设计建议
- 组件库选择建议 (Tailwind, shadcn/ui, etc.)

### 常用前端 MCP Servers

| MCP Server | 用途 | 配置 |
|------------|------|------|
| figma | Figma 设计文件访问 | 已配置 (mcp.json) |
| playwright | 浏览器自动化测试 | mcpServers.playwright |
| context7 | 代码库语义搜索 | mcpServers.context7 |

启动 MCP: `/mcp`

## 禁止行为

- 禁止在未了解代码结构前进行修改
- 禁止在未验证问题的情况下盲目猜测修复
- 禁止删除测试代码来"修复"测试失败
- 禁止使用类型断言抑制类型错误（如 as any、@ts-ignore）
