# docs.cpp 定义文档

## 文件信息
- **源文件**: src/docs.cpp
- **行数范围**: 1-416
- **核心功能**: 文档生成器，生成Odin包的文档

## 类型定义

### 输出格式
```cpp
enum CmdDocFlag {
    CmdDocFlag_Short = 1 << 0,
    CmdDocFlag_InSourceOrder = 1 << 1,
    CmdDocFlag_AllPackages = 1 << 2,
};
```

## 函数签名

### 文档输出
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `print_doc_line(indent, data)` | 打印文档行 | 100 |
| `print_doc_line(indent, fmt, ...)` | 格式化打印 | 108 |
| `print_doc_comment_group_string(indent, g)` | 打印注释组 | 126 |
| `print_doc_expr(expr)` | 打印表达式 | 215 |
| `print_doc_package(info, pkg)` | 打印包文档 | 226 |

### 实体排序
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `cmp_entities_for_printing` | 按类型排序比较 | 31 |
| `cmp_entities_for_printing_by_order_in_src` | 源顺序比较 | 58 |

## 重写建议

### 核心重写
- 输出markdown/html
- Go模板引擎可处理格式化
- 直接打印使用fmt

## 总结
- **整体重写必要性**: 中等
- **建议**: 文档生成逻辑保持，改用Go实现