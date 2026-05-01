# parser.cpp 定义文档

## 文件信息
- **源文件**: src/parser.cpp
- **核心功能**: 语法分析器，将Token序列转换为AST（抽象语法树）

## 类型定义

### AstKind AST节点类型枚举
```cpp
enum AstKind {
    Ast_Invalid,
    // 表达式
    Ast_Ident,
    Ast_Implicit,
    Ast_BasicLit,
    Ast_UnaryExpr,
    Ast_BinaryExpr,
    Ast_CallExpr,
    ast_selector_expr,  // 选择器
    Ast_IndexExpr,
    Ast_SliceExpr,
    Ast_ProcLit,
    // 语句
    Ast_BlockStmt,
    Ast_IfStmt,
    Ast_WhenStmt,
    Ast_SwitchStmt,
    Ast_CaseClause,
    Ast_ForStmt,
    Ast_RangeStmt,
    Ast_ReturnStmt,
    Ast_DeferStmt,
    Ast_BranchStmt,
    Ast_AssignStmt,
    // 声明
    Ast_VariableDecl,
    Ast_ConstantDecl,
    Ast_TypeDecl,
    Ast_ProcDecl,
    // 包级...
    Ast_PackageDecl,
    Ast_ImportDecl,
    Ast_ForeignBlockDecl,
};
```

### AstFile AST文件 (行号: ~前部)
```cpp
struct AstFile {
    CheckerInfo info;
    String fullpath;
    String package_name;
    Array<Ast *> decls;
    // ...
};
```
```go
// 建议Go实现
type AstFile struct {
    Info        CheckerInfo
    FullPath    String
    PackageName String
    Decls       []Ast
}
```

### Ast 节点结构
```cpp
struct Ast {
    AstKind kind;
    TokenPos pos;
    // 根据kind有不同的联合
    // Expression/Statement/Declaration联合
    // ...
    ExactValue builtin;
};
```

### TokenPos 位置信息
```cpp
struct TokenPos {
    i32 file_id;
    i32 offset;
    i32 line;
    i32 column;
};
```

## 函数签名

### AST节点分配
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `ast_node_size(kind)` | 获取节点大小 | Sizeof map | - |
| `alloc_ast_node(f, kind)` | 分配节点 | make/new | - |
| `clone_ast(node, f)` | 克隆节点 | 深度拷贝 | - |
| `clone_ast_array(arr, f)` | 克隆数组 | 遍历克隆 | - |

### 创建表达式
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `ast_bad_expr(f, begin, end)` | 错误表达式 | - |
| `ast_ident(f, token)` | 标识符 | - |
| `ast_basic_lit(f, token)` | 基本字面量 | - |
| `ast_unary_expr(f, op, expr)` | 一元表达式 | - |
| `ast_binary_expr(f, op, l, r)` | 二元表达式 | - |
| `ast_call_expr(f, proc, args, ...)` | 函数调用 | - |
| `ast_selector_expr(f, token, expr, sel)` | 选择表达式 | - |
| `ast_index_expr(f, expr, index, ...)` | 索引表达式 | - |
| `ast_proc_lit(f, type, body, tags, ...)` | 过程字面量 | - |

### 创建语句
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `ast_block_stmt(f, stmts)` | 代码块 | - |
| `ast_if_stmt(f, cond, then, else)` | 条件语句 | - |
| `ast_switch_stmt(f, expr, cases)` | switch语句 | - |
| `ast_for_stmt(f, header, body)` | for循环 | - |
| `ast_return_stmt(f, results)` | 返回语句 | - |
| `ast_defer_stmt(f, stmt)` | defer语句 | - |
| `ast_assign_stmt(f, lhs, rhs, op)` | 赋值语句 | - |

### 创建声明
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `ast_variable_decl(f, name, type, value, ...)` | 变量声明 | - |
| `ast_constant_decl(f, name, type, value, ...)` | 常量声明 | - |
| `ast_type_decl(f, name, type, ...)` | 类型声明 | - |
| `ast_proc_decl(f, type, body, ...)` | 过程声明 | - |

### 值转换
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `exact_value_from_token(f, token)` | Token转 ExactValue | strconv | - |

### 错误处理
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `error(node, fmt, ...)` | 错误报告 | - |
| `error_range(start, end, fmt, ...)` | 范围错误 | - |
| `syntax_error(node, fmt, ...)` | 语法错误 | - |
| `warning(node, fmt, ...)` | 警告 | - |

## 解析函数接口

### 解析器结构
```cpp
struct Parser {
    AstFile *file;
    Token *curr;
    Token *read;  // 预览token
    // ...
};
```

### 解析入口
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `parse_file(path, ...)` | 解析文件 | - |
| `parse_package(f)` | 解析包声明 | - |
| `parse_imports(f)` | 解析import | - |
| `parse_declarations(f)` | 解析声明 | - |

### 各级解析
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `parse_expr(prec)` | 解析表达式 | - |
| `parse_stmt()` | 解析语句 | - |
| `parse_decl()` | 解析声明 | - |
| `parse_type()` | 解析类型 | - |

## 重写建议

### 核心重写要点
1. **AST节点** - 使用Go接口或类型断言
2. **节点内存** - Go GC替代手动管理
3. **Token流** - 可使用迭代器模式

### 可用Go标准库
- **字符串**: `strconv` 用于字面量转换
- **格式化**: `fmt` 用于错误消息

### 重写结构设计
```go
// AST节点接口
type AstNode interface {
    Kind() AstKind
    Pos() TokenPos
}

// 具体表达式
type IdentExpr struct {
    Name String
}
type BinaryExpr struct {
    Op   Token
    Left, Right AstNode
}
// ...
```

### 重写优先级
| 优先级 | 模块 | 重写必要性 |
|--------|------|-----------|
| 高 | AST节点定义 | 需要适配 |
| 高 | 解析函数 | 需要重写 |
| 中 | 错误处理 | 可简化 |
| 低 | 辅助函数 | 较容易 |

## 总结
- **整体重写必要性**: 高
- **建议策略**: 
  - AST节点定义为Go接口+结构体
  - 解析逻辑需要完全重写
  - 使用Go的错误处理简化
  - 可复用tokenizer的结果