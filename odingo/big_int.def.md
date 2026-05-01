# big_int.cpp 定义文档

## 文件信息
- **源文件**: src/big_int.cpp
- **行数范围**: 1-677
- **核心功能**: 大整数运算，基于libtommath库

## 类型定义
```cpp
typedef mp_int BigInt;  // libtommath的大整数类型
```
```go
// 建议Go实现 - 使用math/big
import "math/big"

type BigInt struct {
    *big.Int
}
```

## 函数签名

### 构造函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_from_u64(dst, x)` | 从u64创建 | `new(big.Int).SetUint64(x)` | 149 |
| `big_int_from_i64(dst, x)` | 从i64创建 | `new(big.Int).SetInt64(x)` | 152 |
| `big_int_init(dst, src)` | 初始化复制 | `new(big.Int).Set(x)` | 156 |
| `big_int_from_string(dst, s, success)` | 从字符串 | `new(big.Int).SetString(s, base)` | 186 |

### 转换函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_to_u64(x)` | 转为u64 | `x.Uint64()` | 67 |
| `big_int_to_i64(x)` | 转为i64 | `x.Int64()` | 68 |
| `big_int_to_f64(x)` | 转为f64 | `x.Float64()` | 69 |
| `big_int_to_string(alloc, x, base)` | 转为字符串 | `x.Text(base)` | 70 |

### 算术运算
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_add(dst, x, y)` | 加法 | `z.Add(x, y)` | 72 |
| `big_int_sub(dst, x, y)` | 减法 | `z.Sub(x, y)` | 73 |
| `big_int_mul(dst, x, y)` | 乘法 | `z.Mul(x, y)` | 76 |
| `big_int_quo_rem(x, y, q, r)` | 除法 | `q.DivMod(x, y, r)` | 80 |
| `big_int_shl(dst, x, y)` | 左移 | `z.Lsh(x, n)` | 74 |
| `big_int_shr(dst, x, y)` | 右移 | `z.Rsh(x, n)` | 75 |

### 位运算
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_and(dst, x, y)` | 与 | `z.And(x, y)` | 84 |
| `big_int_or(dst, x, y)` | 或 | `z.Or(x, y)` | 87 |
| `big_int_xor(dst, x, y)` | 异或 | `z.Xor(x, y)` | 86 |
| `big_int_not(dst, x, count, signed)` | 取反 | `z.Not(x)` | 88 |

### 符号操作
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_is_neg(x)` | 是否负数 | `x.Sign() < 0` | 100 |
| `big_int_neg(dst, x)` | 取负 | `z.Neg(x)` | 101 |
| `big_int_sign(x)` | 获取符号 | `x.Sign()` | 141 |

### 内存管理
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `big_int_dealloc(dst)` | 释放 | GC自动 | 56 |

## 重写建议

### 核心重写要点
1. **使用math/big** - Go标准库的大整数实现
2. **语义兼容** - big.Int方法和mp_int略有不同

### 直接映射
```go
import "math/big"

// 包装
type BigInt big.Int

func (b *BigInt) Add(x, y *BigInt) *BigInt {
    r := new(big.Int)
    r.Add(&x.Int, &y.Int)
    return (*BigInt)(r)
}
```

## 总结
- **整体重写必要性**: 低
- **建议策略**: 
  - 使用 `math/big` 包
  - 包装为自定义类型以兼容接口
  - 大部分操作可直接映射