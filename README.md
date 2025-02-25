# MySQL Binlog Listener

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/mysql-binlog-listener)](https://goreportcard.com/report/github.com/yourusername/mysql-binlog-listener)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight MySQL binlog listener with Redis position persistence, written in Go.

## 📌 Features

- ✅ Real-time monitoring of INSERT/UPDATE/DELETE events
- 🚀 Automatic schema mapping with struct tags
- 🔄 Redis-based position recovery
- 🛠 Extensible position storage interface
- 🧩 Built-in error handling with channel
- 🔌 Easy integration with existing systems

## 📦 Installation

```bash
go get github.com/YuanQianJ/binlog/master/binlog
