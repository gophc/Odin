# exact_value.cpp 定义文档

## 文件信息
- **源文件**: src/exact_value.cpp
- **核心功能**: 精确值表示，用于编译时常量

## 类型定义

### ExactValue 结构
```cpp
enum ExactValueKind {
    ExactValue_Invalid,
    ExactValue_Bool,
    ExactValue_Integer,
    ExactValue_Float,
    ExactValue_Complex,
    ExactValue_String,
    ExactValue_Pointer,
    ExactValue_TypeId,
    ExactValue_Array,
    ExactValue_Slice,
    ExactValue_Map,
    ExactValue_Struct,
    ExactValue_Union,
};

struct ExactValue {
    ExactValueKind kind;
    union {
        bool value_bool;
        BigInt value_integer;
        f64 value_float;
        // ...
    };
};
```
```go
// 建议Go实现
type ExactValueKind int

type ExactValue struct {
    Kind  ExactValueKind
    Value interface{}  // 使用interface{}存储
}
```

## 函数签名

### 构造函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `exact_value_bool(b)` | 布尔值 | 直接构造 | - |
| `exact_value_integer(x)` | 整数值 | SetInt64 | - |
| `exact_value_float(f)` | 浮点值 | 设置float | - |
| `exact_value_string(s)` | 字符串 | 设置string | - |

### 转换函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `exact_value_integer_from_string(s)` | 字符串→整数 | strconv.ParseInt | - |
| `exact_value_float_from_string(s)` | 字符串→浮点 | strconv.ParseFloat | - |

## 重写建议

### 直接映射
- 大部分操作可直接用interface{}实现

## 总结
- **整体重写必要性**: 低
- **建议**: 使用interface{}简化