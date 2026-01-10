# Contributing to EOCall

感谢你对 EOCall 项目的关注！我们欢迎各种形式的贡献。

## 开发环境

### 前置要求

- Go 1.23+
- Make

### 本地开发

```bash
# 克隆项目
git clone https://github.com/NuyoahCh/eocall.git
cd eocall

# 安装依赖
go mod download

# 运行测试
make test

# 构建
make build
```

## 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `test`: 测试相关
- `refactor`: 代码重构
- `chore`: 构建/工具相关

示例：
```
feat: add streaming response support
fix: resolve session isolation issue
docs: update README with examples
```

## Pull Request 流程

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 代码规范

- 使用 `gofmt` 格式化代码
- 通过 `golangci-lint` 检查
- 为新功能添加测试
- 保持代码简洁清晰

## 问题反馈

如果你发现 Bug 或有功能建议，请创建 Issue 并提供：

- 问题描述
- 复现步骤
- 期望行为
- 环境信息 (Go 版本、操作系统等)

## License

贡献的代码将采用 MIT License。
