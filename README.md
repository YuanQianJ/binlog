# binlog
binlog 通用监听框架

Go Report Card
License: MIT

一个基于Go语言的MySQL binlog监听工具，提供自动化的数据变更捕获和Redis位置持久化功能，支持实时同步数据库变更到其他系统。

特性
🔍 基于go-mysql库的高效binlog解析

🚀 支持INSERT/UPDATE/DELETE事件监听

🛠️ 自动结构体映射（支持自定义标签）

📌 Redis持久化binlog位置（断点续传）

🧩 可扩展的位置存储实现（支持自定义存储）

🔒 线程安全的错误处理机制

快速开始
安装
go get github.com/YuanQianJ/binlog
