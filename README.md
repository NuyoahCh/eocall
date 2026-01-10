# EOCall - OnCall 响应引擎

基于 Go 语言开发的智能 OnCall 响应引擎，集成知识库、对话与运维能力，通过 AI Agent 实现故障秒级应答与根因分析。

## 特性

- 🔧 **工具链编排** - 标准化工具协议，封装日志监控查询能力，LLM 意图识别与自主编排
- 📚 **知识库检索增强 (RAG)** - 基于向量数据库与语义匹配的高精度智能问答
- 🤖 **智能体决策闭环** - Plan-Execute 模式，从告警分析到工具执行的自动化闭环
- ⚡ **流式响应** - Streaming 流式输出，消除大模型生成延迟
- 🔒 **并发会话隔离** - 基于 UserID 的会话隔离，支持百级并发

## 技术栈

- Go 1.23+
- [Eino](https://github.com/cloudwego/eino) - ByteDance 大语言模型应用框架
- 向量数据库 (可配置)

## 项目结构

```
eocall/
├── cmd/                    # 应用入口
│   └── server/            # HTTP/gRPC 服务
├── internal/              # 内部包
│   ├── agent/             # AI Agent 核心
│   │   ├── planner/       # Plan-Execute 规划器
│   │   └── executor/      # 工具执行器
│   ├── chat/              # 对话管理
│   │   ├── session/       # 会话隔离
│   │   └── memory/        # 上下文记忆
│   ├── rag/               # RAG 检索增强
│   │   ├── embedding/     # 向量嵌入
│   │   ├── retriever/     # 检索器
│   │   └── reranker/      # 重排序
│   ├── tools/             # 工具链
│   │   ├── registry/      # 工具注册
│   │   ├── log/           # 日志查询工具
│   │   └── monitor/       # 监控查询工具
│   └── llm/               # LLM 适配层
├── pkg/                   # 公共包
│   ├── config/            # 配置管理
│   ├── logger/            # 日志
│   └── errors/            # 错误定义
├── api/                   # API 定义
│   └── proto/             # Protobuf 定义
└── configs/               # 配置文件
```

## 快速开始

```bash
# 克隆项目
git clone https://github.com/NuyoahCh/eocall.git
cd eocall

# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

## 配置

复制配置模板并修改：

```bash
cp configs/config.example.yaml configs/config.yaml
```

## License

MIT License
