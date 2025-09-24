# Go-Ethereum 核心功能与架构设计研究

## 项目概述
本项目是对以太坊参考实现 Go-Ethereum（Geth）的深入研究，通过理论分析、架构设计和实践验证三个维度，全面理解区块链核心组件的实现原理。

## 项目结构
```
task02/
├── README.md                    # 项目说明
├── TASK02.md                   # 原始任务要求
├── docs/                       # 文档目录
│   ├── 01-理论分析.md           # Geth定位与核心模块分析
│   ├── 02-架构设计.md           # 分层架构设计文档
│   ├── 03-实践验证.md           # 实践操作记录
│   └── 04-最终报告.md           # 综合研究报告
├── diagrams/                   # 架构图表
│   ├── architecture.puml       # PlantUML架构图源码
│   ├── transaction-flow.puml   # 交易流程图
│   └── state-model.puml        # 状态存储模型
├── scripts/                    # 实践脚本
│   ├── setup-geth.sh          # Geth环境搭建
│   ├── deploy-contract.js      # 智能合约部署
│   └── test-functions.js       # 功能测试脚本
└── evidence/                   # 实践证据
    ├── screenshots/            # 操作截图
    ├── logs/                   # 运行日志
    └── contracts/              # 测试合约
```

## 任务完成情况
- [x] 项目结构搭建
- [ ] 理论分析（40%）
- [ ] 架构设计（30%）
- [ ] 实践验证（30%）
- [ ] 最终报告整理

## 评分标准
- 架构完整性：40%
- 实现深度：30%
- 实践完成度：30%

## 开始使用
1. 查看理论分析：`docs/01-理论分析.md`
2. 查看架构设计：`docs/02-架构设计.md`
3. 运行实践验证：`scripts/setup-geth.sh`
4. 查看最终报告：`docs/04-最终报告.md`