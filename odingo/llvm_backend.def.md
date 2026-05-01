# llvm_backend.cpp 定义文档

## 文件信息
- **源文件**: src/llvm_backend.cpp (以及相关文件llvm_backend_*.cpp)
- **核心功能**: LLVM后端，将AST转换为LLVM IR并生成目标机器码

## 类型定义

### lbModule LLVM模块
```cpp
struct lbModule {
    // LLVM模块上下文
    LLVMModuleRef module;
    LLVMExecutionEngineRef engine;
    // 代码生成数据...
};
```

### lbProcedure LLVM过程
```cpp
struct lbProcedure {
    lbModule *m;
    LLVMValueRef value;
    Array<lbBlock> blocks;
    lbBlock *curr_block;
    // 参数和局部变量...
};
```

### lbValue LLVM值
```cpp
struct lbValue {
    LLVMValueRef value;
    Type *type;
};
```

### lbBlock 基本块
```cpp
struct lbBlock {
    LLVMBasicBlockRef block;
    Array<lbInstruction> insts;
    lbProcedure *proc;
};
```

## 函数签名

### 模块初始化
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `lb_add_foreign_library_path(m, e)` | 添加外部库 | - |
| `lb_correct_entity_linkage(gen)` | 修正链接 | - |

### 代码生成
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `lb_equal_proc_for_type(m, type)` | 生成相等比较过程 | - |
| `lb_hasher_proc_for_type(m, type)` | 生成哈希过程 | - |
| `lb_map_get_proc_for_type(m, type)` | 生成Map获取 | - |
| `lb_map_set_proc_for_type(m, type)` | 生成Map设置 | - |

### 上下文管理
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `lb_push_context_onto_stack(p, ctx)` | 推入上下文 | - |
| `lb_push_context_onto_stack_from_implicit_parameter(p)` | 从参数获取上下文 | - |

### 优化相关
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `lb_add_callsite_force_inline(p, ret)` | 强制内联 | - |

## 重写建议

### 核心重写要点
1. **LLVM集成** - 使用llvm.org/go-bindings或CGO调用
2. **IR构建** - 使用LLVM C API
3. **代码生成** - 调用LLVM工具链

### 方案选择

#### 方案1: 使用Go+LLVM绑定
```go
/*
利用LLVM的Go绑定或CGO
- 优点: 直接使用LLVM IR
- 缺点: 需要绑定实现
*/
import "llvm"
```

#### 方案2: 使用子进程调用odin+llc
```go
/*
作为子进程调用现有odin编译器
- 优点: 可复用现有实现
- 缺点: 进程间通信开销
*/
func GenerateLLVMIR(ast *AstFile) ([]byte, error) {
    // 序列化AST，调用现有odin
}
```

#### 方案3: 生成C代码
```go
/*
生成C代码，使用gcc/clang编译
- 优点: 无需LLVM依赖
- 缺点: 功能受限
*/
```

### 重写优先级
| 优先级 | 模块 | 重写必要性 |
|--------|------|-----------|
| 高 | LLVM调用层 | 需选择方案 |
| 中 | IR生成 | 可复用 |
| 中 | 优化passes | 可调用LLVM工具 |

## 总结
- **整体重写必要性**: 高
- **建议策略**: 
  - 选项1: 使用CGO调用现有odin或LLVM工具
  - 选项2: 生成C代码使用gcc编译
  - 建议使用子进程方案保持兼容性