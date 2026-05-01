# common.cpp 定义文档

## 文件信息
- **源文件**: src/common.cpp
- **行数范围**: 1-928
- **核心功能**: 提供通用基础工具函数、数据结构和内存操作

## 类型定义

### 基础类型别名 (使用Go内建可实现)
| C++类型 | Go等价 | 重写必要性 |
|--------|-------|-----------|
| `i32, i64, u32, u64` | `int32, int64, uint32, uint64` | **无需重写** - 使用Go内置类型 |
| `isize, usize` | `int, uint` | **无需重写** - Go的int/uint |
| `f32, f64` | `float32, float64` | **无需重写** - Go内置 |
| `bool` | `bool` | **无需重写** |
| `Rune` | `rune` | **无需重写** |

### 字符串类型
| C++类型 | Go等价 | 重写必要性 |
|--------|-------|-----------|
| `String` | `struct { text []byte; len int }` | **需适配** - 参考strings.Builder |

### 动态数组
| C++类型 | Go等价 | 重写必要性 |
|--------|-------|-----------|
| `Array<T>` | `[]T` | **需适配** - Go slice语义略有不同 |

### 需要重写的类型
```cpp
// 核心数据结构需要Go重写
struct String { u8 *text; isize len; };  // → type String struct{ Text []byte; Len int }
struct gbString {};                     // → 需要字符串构建器
```

## 函数签名

### 工具函数 - 可用Go内建实现
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `is_power_of_two(i64)` | 检查2的幂 | 使用 `x&(x-1) == 0` | 83 |
| `next_pow2(i32/i64)` | 向上取整2的幂 | bit操作实现 | 394-420 |
| `bit_set_count(u32/u64)` | 统计置位数 | 使用 `bits.OnesCount` | 452-466 |
| `floor_log2(u32/u64)` | 向下取整对数 | `bits.Len(x)-1` | 468-485 |
| `ceil_log2(u32/u64)` | 向上取整对数 | `64-bits.Len(x-1)` | 488-507 |

### 哈希函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `fnv32a(data, len)` | FNV哈希 | 使用 `fnv` 包 | 130 |
| `fnv64a(data, len, seed)` | FNV64哈希 | `hash/fnv` 包 | 151 |

### 字符串转换
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `u64_from_string(String)` | 字符串转u64 | `strconv.ParseUint` | 201 |
| `u64_to_string(u64, buf, len)` | u64转字符串 | `strconv.FormatUint` | 244 |
| `i64_to_string(i64, buf, len)` | i64转字符串 | `strconv.FormatInt` | 259 |

### 整数运算 - 可用Go实现
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `add_overflow_u64(x, y, *result)` | 加法溢出检测 | 用 `^uint` 绕过 | 321 |
| `sub_overflow_u64(x, y, *result)` | 减法溢出检测 | 用 `^uint` 绕过 | 326 |
| `mul_overflow_u64(x, y, *lo, *hi)` | 乘法产生128位 | `bits.Mul64` | 331 |

## 数据结构 - 需要重写

### String类
```cpp
// C++
struct String { u8 *text; isize len; };
```
```go
// 建议Go实现
type String struct {
    Text []byte
    Len  int
}

// 常用操作封装
func (s String) Slice(start, end int) String { ... }
func (s String) Equal(other String) bool { ... }
```

### 数据结构容器
- `PtrMap` - 指针键map → 使用 `map[uintptr]any`
- `PtrSet` - 指针集合 → 使用 `map[uintptr]struct{}`
- `StringMap` - 字符串键map → 使用 `map[string]Value`
- `StringSet` - 字符串集合 → 使用 `map[string]struct{}`
- `String16Map` - u16键map → 使用 `map[uint16]Value`
- `PriorityQueue` → 使用 `container/heap`

## 重写建议

### 可完全用Go内置替代
1. **数值函数**: 全部使用Go标准库 `math`, `bits`, `strconv`
2. **基础类型**: 直接用Go内置类型
3. **字符串**: 使用 `strings.Builder` 或 自定义 String

### 需要自行实现
1. **自定义分配器**: Go的GC机制可处理
2. **线程池**: 使用 `sync.Pool` + goroutine
3. **原子操作**: 使用 `sync/atomic`
4. **字符串interning**: 使用 `string.Intern` 或 map

## 总结
- **整体重写必要性**: 中等
- **建议策略**: 
  - 多数工具函数可用Go标准库替代
  - 核心数据结构需要适配Go的slice/map语义
  - 分配器相关代码可利用Go GC简化