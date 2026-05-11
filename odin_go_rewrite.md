# Odin编译器 Go重写计划

## 项目概述

 Odin是一个用C++编写的编译器，现计划用Go重写。本文档基于入口文件 src/main.cpp 的完整逻辑链条制定重写计划。

---

## 入口 main() 完整调用链 (src/main.cpp:4071-4269)

```
main()
│
├─ 初始化阶段
│  ├─ init_thread_pool()              行4059
│  └─ init_universal()            行4064
│
├─ 解析阶段 [TIME_SECTION: parse files]
│  ├─ init_parser()              行4073
│  └─ parse_packages()        行4080
│     └─ tokenizer.cpp, parser.cpp
│
├─ 类型检查阶段 [TIME_SECTION: type check]
│  ├─ init_checker()          行4091
│  └─ check_parsed_files()    行4102
│     └─ checker.cpp, check_*.cpp
│
├─ 文档生成 [TIME_SECTION: generate documentation]
│  └─ generate_documentation()  行4137
│     └─ docs.cpp
│
├─ 代码生成阶段 [TIME_SECTION: LLVM API Code Gen]
│  ├─ lb_init_generator()     行4177
│  └─ lb_generate_code()    行4186
│     └─ llvm_backend.cpp, llvm_backend_*.cpp
│
├─ 链接阶段
│  └─ linker_stage()        行4191
│     └─ linker.cpp, bundle_command.cpp
│
└─ 执行阶段 (run命令)
   └─ system_must_exec_command_line_app()
```

---

## 模块重写必要性分级

| 模块 | 负责阶段 | 重写必要性 | 原因 |
|------|---------|-----------|------|
| **common.cpp** | 基础设施 | 中等 | 基础工具函数，Go标准库可替代大部分 |
| **tokenizer.cpp** | 解析 | **高** | 词法分析需要重写 |
| **parser.cpp** | 解析 | **高** | AST解析需要完全重写 |
| **types.cpp** | 解析/检查 | 中等高 | 类型系统需适配Go泛型 |
| **checker.cpp** | 类型检查 | 中等 | 语义检查需适配 |
| **check_*.cpp** | 类型检查 | 中等 | 类型检查逻辑 |
| **docs.cpp** | 文档 | 中等 | 文档生成可简化 |
| **llvm_backend.cpp** | 代码生成 | **高** | LLVM调用需选方案 |
| **linker.cpp** | 链接 | 低 | 外部调用可复用 |
| **bundle_command.cpp** | 链接 | 低 | 打包逻辑 |
| **build_settings.cpp** | 初始化 | 低 | 配置解析可简化 |

---

## 可用Go内建功能实现的部分

### 无需深入重写 (直接使用Go标准库)
| C++模块 | 功能 | Go建议 | 使用位置 |
|---------|------|-------|--------|
| 数值运算 | 加减乘除 | `math`, `bits` 包 | common.cpp |
| 大整数 | big.Int | `math/big` 包 | big_int.cpp |
| 字符串转换 | ParseInt/FormatInt | `strconv` 包 | common.cpp, exact_value.cpp |
| 哈希 | FNV | `hash/fnv` 包 | common.cpp |
| Unicode | 字符处理 | `unicode`/`unicode/utf8` 包 | tokenizer.cpp |
| 线程池 | Pool + goroutine | `sync.Pool` + `go func()` | main.cpp thread_pool |
| 原子操作 | atomic | `sync/atomic` 包 | 各模块atomic字段 |
| 堆 | 优先级队列 | `container/heap` | priority_queue.cpp |

### 需要适配 (结构差异)
| C++结构 | 问题 | 适配方案 |
|---------|------|---------|
| `String` (u8*, len) | Go strings不可变 | 包装 `type String struct { Text []byte }` |
| `Array<T>` | 初始化差异 | 用 `[]T` 替代 |
| `Map<K,V>` | 内部实现差异 | 直接用 `map[K]V` |
| 手动内存分配 | GC替代 | 放弃手动管理 |

---

## 详细重写顺序和计划

### 阶段1: 基础设施 (Week 1-2)

#### 1.1 类型系统基础
```
创建文件:
- internal/types/types.go        - 类型定义
- internal/common/string.go   - String包装
- internal/common/arena.go   - 内存池(可选)

核心类型映射:
- i32,i64,u32,u64    → int32,int64,uint32,uint64
- isize,usize       → int,uint  
- f32,f64         → float32,float64
- bool             → bool
- String           → struct { Text []byte }
```

#### 1.2 工具函数迁移
```
创建文件:
- internal/common/util.go    - 位操作、数学
- internal/common/hash.go - 哈希函数
- internal/container/   - 容器(多数可用标准库)

标准库直接替代:
- bits.OnesCount    → bit_set_count
- bits.Len        → floor_log2
- math/big.Int    → BigInt
- hash/fnv.New64  → fnv64a
```

### 阶段2: 词法分析器 (Week 2-3)

#### 2.1 Tokenizer (高优先级)
```
目标: 将源代码转换为Token序列

创建文件:
- internal/tokenizer/token.go   - Token定义
- internal/tokenizer/lexer.go  - Lexer实现

关键挑战:
- 位置追踪(Line/Column维护) → 内部状态
- Unicode处理           → unicode/utf8.DecodeRune
- Token生成逻辑        → 逐字符分析

建议:
- 使用map[string]TokenKind替代关键字哈希表
- 实现Peek()方法用于预读
```

#### 2.2 错误处理
```
创建文件:
- internal/error/error.go  - 错误报告

建议:
- 统一使用error而非异常
- 位置信息嵌入错误
```

### 阶段3: 语法分析器 (Week 3-4)

#### 3.1 AST定义
```
目标: 将Token序列转换为AST

创建文件:
- internal/parser/ast.go     - AST节点接口
- internal/parser/expr.go  - 表达式解析
- internal/parser/stmt.go   - 语句解析
- internal/parser/decl.go  - 声明解析
- internal/parser/type.go - 类型解析

关键挑战:
- 递归下降解析
- AST节点内存管理 → GC处理
- 优先级表达式 → 运算符优先级表

建议接口:
type Node interface {
    Kind() AstKind
    Pos() TokenPos
}
```

#### 3.2 解析器主逻辑
```
func ParseFile(path string) (*AstFile, error)
func ParsePackage(...)
func ParseExpr(...)
```

### 阶段4: 类型检查器 (Week 4-5)

#### 4.1 作用域管理
```
创建文件:
- internal/checker/scope.go   - 作用域
- internal/checker/entity.go - 实体定义

建议结构:
type Scope struct {
    Parent  *Scope
    Elements map[string]*Entity
}

func (c *Checker) EnterScope()
func (c *Checker) ExitScope()
func (c *Checker) Lookup(name string) *Entity
```

#### 4.2 类型检查
```
创建文件:
- internal/checker/check.go   - 检查器主逻辑
- internal/checker/resolve.go - 符号解析
- internal/checker/vet.go      - vet检查

核心:
- 类型推断
- 表达式类型检查
- 兼容性检查
```

### 阶段5: 代码生成 (Week 5-7)

#### 5.1 方案选择

**方案A(推荐): 子进程调用现有编译器**
```
优点:
- 可快速实现
- 复用现有LLVM后端
- 降低风险

实现:
- Go负责前端(Parse/Check)
- 调用现有odin编译器的linker_stage
```

**方案B: Go + CGO调用LLVM**
```
优点:
- 完全Go实现

缺点:
- LLVM绑定复杂性
- 跨平台编译难度
```

**方案C: 生成C代码**
```
优点:
- 无LLVM依赖

缺点:
- 功能受限
- 性能损失
```

#### 5.2 链接器包装
```
创建文件:
- internal/linker/command.go - 参数构建
- internal/linker/exec.go   - 调用链接器

建议: 使用os/exec包
```

### 阶段6: 命令行接口 (Week 7)

```
创建文件:
- cmd/odingo/main.go  - 主入口

需要处理:
- 命令解析 (build/run/test/check/doc)
- 标志解析 (100+标志)
- 用法输出
```

---

## 技术决策

### 1. 泛型使用 (Go 1.18+)
```go
type Type[T any] struct {
    Value T
}
```

### 2. 错误处理
```go
func Parse(src string) (*AstFile, error)  // 而非异常
```

### 3. 并发模型
```go
// 线程池使用goroutine
type Pool struct {
    workers int
    jobs    chan Job
}
```

### 4. 内存管理
放弃手动分配，使用Go GC

---

## 详细实施时间表

| Week | 任务 | 产出 |
|------|------|------|
| 1 | 基础类型和String | internal/types, common |
| 2 | 工具函数 | util, hash, container |
| 3 | Tokenizer | lexer, token |
| 4 | Parser基础 | AST, 解析器框架 |
| 5 | 表达式解析 | expr解析 |
| 6 | 类型检查器 | checker框架 |
| 7 | 链接器 | 命令行，链接 |
| 8 | 测试和优化 | 集成测试 |

---

## 风险和缓解

### 风险1: LLVM依赖
**缓解**: 初期使用子进程调用现有编译器(方案A)

### 风险2: 性能差距
**缓解**: 
- Profiling定位热点
- 关键路径用优化过的Go代码

### 风险3: 100+编译标志
**缓解**: 
- flag/pflag包处理
- 分模块定义

---

## 推荐包结构

```
odingo/
├── cmd/
│   └── odin/
│       └── main.go
├── internal/
│   ├── common/           # 基础工具
│   │   ├── string.go
│   │   ├── util.go
│   │   └── pool.go
│   ├── tokenizer/        # 词法分析
│   │   ├── token.go
│   │   └── lexer.go
│   ├── parser/         # 语法分析
│   │   ├── ast.go
│   │   ├── expr.go
│   │   ├── stmt.go
│   │   └── parser.go
│   ├── checker/        # 语义分析
│   │   ├── scope.go
│   │   ├── entity.go
│   │   └── check.go
│   ├── types/         # 类型系统
│   │   └── types.go
│   ├── linker/        # 链接
│   │   └── linker.go
│   └── error/         # 错误处理
│       └── error.go
├── doc/              # 文档生成
│   └── doc.go
└── config/          # 配置
    └── config.go
```

---

## 总结

**重写可行性**: 高

**预计工作量**: 6-8人月

**推荐策略**: 
1. 阶段1-4: 实现Go前端(Parse/Check)
2. 阶段5: 方案A子进程调用现有后端
3. 阶段6: 整合测试

**关键成功因素**:
- 保持命令行参数兼容
- 详细测试验证(golden test)
- 性能优化迭代