# DataSeap - Unified Data Foundation Service for DIB

[![Go Report Card](https://goreportcard.com/badge/github.com/turtacn/DataSeap)](https://goreportcard.com/report/github.com/turtacn/DataSeap)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org/dl/)
[![GitHub issues](https://img.shields.io/github/issues/turtacn/DataSeap)](https://github.com/turtacn/DataSeap/issues)


**DataSeap (Data Sea Platform)** 是一个旨在统一、高效、可扩展数据湖底座的开源项目。它基于业界领先的开源数据技术（如 StarRocks [1], Apache Pulsar等），封装和集成了对底层大数据集群的监控、告警、日志、事件、升级、扩缩容等维护能力，并提供标准化的数据上报、查询与分析接口。(Beyond the concept of data lake)

**核心目标**: 简化业务对底层数据湖构建和运维的复杂度。通过汇聚多源异构的数据，利用先进的分析引擎和AI能力，实现并保障数据资产的安全、合规流转与价值释放。

**[English Version](./README_EN.md)** (To be created)

## 项目背景与愿景

需要处理海量的、多源异构的数据，并进行实时分析、深度挖掘和历史回溯。传统的数据架构往往难以应对这些挑战，导致查询性能瓶颈、运维复杂、成本高昂等问题。

DataSeap 致力于：
* **统一数据底座**: 为上层安全应用提供一个统一的数据存储、查询和分析入口。
* **性能卓越**: 基于 StarRocks 的极速分析能力，满足实时查询和复杂分析的性能要求。
* **弹性伸缩**: 架构设计支持水平扩展，从容应对数据量和业务增长。
* **运维简化**: 提供集成的运维管理能力，降低大数据平台的运维门槛。
* **开放融合**: 基于开源技术栈，易于集成和扩展，促进安全生态的协同发展。

## 主要特性

* **统一数据接口**:
    * 标准化的数据上报API，支持多种数据源接入。
    * 统一的查询API，支持实时点查、聚合分析、日志检索、全文检索等。
* **高性能分析引擎**:
    * 深度整合 StarRocks，利用其MPP架构、向量化执行引擎、CBO优化器、物化视图、多维索引等特性。
    * 支持高效的全文检索，可跨表查询并返回详细的匹配信息（表名、字段名）。
* **工作负载隔离**:
    * 基于 StarRocks Workload Group 实现多场景任务（查询、分析、导入）的资源隔离和优先级调度。
* **数据管理**:
    * 辅助管理 StarRocks 中的表结构、分区、分桶、索引（包括倒排索引）和物化视图。
* **平台运维**:
    * 集成了对 StarRocks 集群、Pulsar 集群等的监控、告警、日志收集、事件追踪能力。
    * 提供集群升级、扩缩容等生命周期管理的接口或脚本。
* **多语言分词器支持**:
    * 在全文检索场景中，支持 standard, english, chinese 等多种分词器。

## 架构概览

DataSeap 采用分层架构，主要包括展现与接入层、应用服务层、数据抽象与访问层、数据平台层以及运维管理平台。

<img src="docs/imgs/overview-arch.png" width="100%"/>

更详细的架构设计请参见: **[architecture.md](./docs/architecture.md)**

## 技术栈

  * **核心数据引擎**: [StarRocks](https://github.com/StarRocks/starrocks)
  * **消息队列**: [Apache Pulsar](https://pulsar.apache.org/)
  * **后端开发**: Go (\>=1.20)
  * **配置管理**: Viper (计划)
  * **API框架**: Gin (计划)
  * **ORM/数据库驱动**: GORM (用于内部元数据，可选), StarRocks Go Driver (计划)
  * **日志**: Zap (计划)
  * **监控**: Prometheus, Grafana (通过导出指标)
  * **容器化**: Docker, Kubernetes (推荐部署方式)

## 快速开始

### 前提条件

  * Go \>= 1.20
  * Git
  * Docker 和 Docker Compose (用于本地快速启动依赖服务)
  * (可选) Kubernetes 集群

### 安装与构建

1.  **克隆代码库**:

    ```bash
    git clone [https://github.com/turtacn/DataSeap.git](https://github.com/turtacn/DataSeap.git)
    cd DataSeap
    ```

2.  **构建 DataSeap 服务**:
    (详细构建脚本待 `scripts/build.sh` 完成后提供)

    ```bash
    # 示例 (具体命令待定)
    # go build -o build/DataSeap_server ./cmd/DataSeap-server
    ./scripts/build.sh
    ```

### 运行

#### 1\. 启动依赖服务 (StarRocks, Pulsar)

项目后续会提供 `docker-compose.yaml` 文件，用于在本地快速启动 StarRocks 和 Pulsar 集群以供开发和测试。

```bash
# (示例，待 docker-compose 文件提供后更新)
# docker-compose -f deployments/docker-compose/dev-env.yml up -d
```

#### 2\. 运行 DataSeap 服务

(详细运行脚本待 `scripts/run.sh` 完成后提供)

```bash
# 示例 (具体命令待定)
# ./build/DataSeap_server --config=./configs/config.yaml
./scripts/run.sh
```

启动成功后，DataSeap 服务将在配置文件中指定的端口上监听请求 (例如 `http://localhost:8080`)。

## 开发

### 代码结构

项目代码 `(Brief)` 目录下，遵循典型的分层架构：

```text
dataseap/
├── api/               # API definitions (gRPC protos, OpenAPI specs)
├── cmd/               # Main application(s)
├── deployments/       # Dockerfiles, Kubernetes manifests, etc.
├── docs/              # Project documentation
├── internal/          # Private application and library code
├── pkg/               # Public library code, organized by layer/feature
│   ├── adapter/       # Adapters to external systems (DBs, queues)
│   ├── common/        # Shared utilities, types, constants, errors
│   ├── domain/        # Core business logic, entities, service interfaces
│   ├── observability/ # Tracing, metrics collection setup
│   └── transport/     # gRPC and HTTP/REST transport layers
├── scripts/           # Build, lint, test scripts
├── tests/             # Integration and E2E tests
├── go.mod
└── README.md
```

### 编码规范

  * 遵循 Go 官方的编码风格指南。
  * 所有公开的函数和类型都需要有清晰的英文注释。
  * 鼓励编写单元测试和集成测试。

## 贡献指南

我们欢迎任何形式的贡献！无论是代码提交、问题反馈、文档改进还是功能建议。

1.  **Fork 本仓库**
2.  **创建您的特性分支**: `git checkout -b feature/AmazingFeature`
3.  **提交您的更改**: `git commit -m 'Add some AmazingFeature'`
      * 请确保您的提交信息清晰明了，遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范更佳。
4.  **将更改推送到分支**: `git push origin feature/AmazingFeature`
5.  **开启一个 Pull Request**

在提交 Pull Request 之前，请确保您的代码：

  * 通过了所有测试 (`go test ./...`)。
  * 遵循了项目的编码规范。
  * (如果适用) 更新了相关文档。

## 许可证

本项目采用 [Apache License 2.0](https://www.google.com/search?q=./LICENSE) 许可证。

## 联系方式

  * **GitHub Issues**: [https://www.google.com/url?sa=E\&source=gmail\&q=https://github.com/turtacn/DataSeap/issues](https://www.google.com/url?sa=E&source=gmail&q=https://github.com/turtacn/DataSeap/issues)
  * (后续可添加其他联系方式，如邮件列表、社区论坛等)

-----

**参考资料**

- [1] StarRocks Project. *The world's fastest open query engine for sub-second analytics both on and off the data lakehouse.* GitHub. [https://github.com/StarRocks/starrocks](https://github.com/StarRocks/starrocks)
