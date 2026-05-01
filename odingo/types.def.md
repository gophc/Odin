# types.cpp 定义文档

## 文件信息
- **源文件**: src/types.cpp
- **行数范围**: 1-5603
- **核心功能**: Odin类型系统核心定义，包含所有基础类型和复合类型

## 类型定义

### BasicKind 基础类型枚举 (行号: 8-99)
```cpp
enum BasicKind {
    Basic_Invalid,
    Basic_bool, Basic_b8, Basic_b16, Basic_b32, Basic_b64,
    Basic_i8, Basic_u8, Basic_i16, Basic_u16, Basic_i32, Basic_u32, Basic_i64, Basic_u64, Basic_i128, Basic_u128,
    Basic_rune,
    Basic_f16, Basic_f32, Basic_f64,
    Basic_complex32, Basic_complex64, Basic_complex128,
    Basic_quaternion64, Basic_quaternion128, Basic_quaternion256,
    Basic_int, Basic_uint, Basic_uintptr, Basic_rawptr,
    Basic_string, Basic_cstring,
    Basic_string16, Basic_cstring16,
    Basic_any, Basic_typeid,
    // 字节序相关...
    Basic_i16le, Basic_u16le, ... Basic_f64le,
    Basic_i16be, Basic_u16be, ... Basic_f64be,
    // 未类型化...
    Basic_UntypedBool, Basic_UntypedInteger, Basic_UntypedFloat, ...
};
```
```go
// 建议Go实现 - 直接用enum
type BasicKind int

const (
    BasicInvalid BasicKind = iota
    BasicBool
    BasicB8, BasicB16, BasicB32, BasicB64
    BasicI8, BasicU8, BasicI16, BasicU16, BasicI32, BasicU32, BasicI64, BasicU64, BasicI128, BasicU128
    BasicRune
    BasicF16, BasicF32, BasicF64
    // ...
    BasicInt, BasicUint, BasicUintptr, BasicRawptr
    BasicString, BasicCstring
    // ...
)
```

### TypeKind 类型类别枚举 (行号: 318-324)
```cpp
enum TypeKind {
    Type_Invalid,
    Type_Basic,
    Type_Named,
    Type_Generic,
    Type_Pointer,
    Type_MultiPointer,
    Type_Array,
    Type_EnumeratedArray,
    Type_Slice,
    Type_DynamicArray,
    Type_FixedCapacityDynamicArray,
    Type_Map,
    Type_Struct,
    Type_Union,
    Type_Enum,
    Type_Tuple,
    Type_Proc,
    Type_BitSet,
    Type_SimdVector,
    Type_Matrix,
    Type_BitField,
    Type_SoaPointer,
    Type_Count,
};
```

### 核心类型结构

#### Type 基本结构 (行号: 343-357)
```cpp
struct Type {
    TypeKind kind;
    union {
        // 根据kind选择
        BasicType Basic;
        TypeNamed Named;
        // ...其他类型
    };
    std::atomic<i64> cached_size;
    std::atomic<i64> cached_align;
    std::atomic<u64> canonical_hash;
    std::atomic<u32> flags;
    bool failure;
};
```
```go
// 建议Go实现
type Type struct {
    Kind        TypeKind
    // 使用空接口或类型断言
    Value      interface{}  // 具体类型值
    CachedSize int64
    CachedAlign int64
    CanonicalHash uint64
    Flags       uint32
    Failure     bool
}
```

#### TypeStruct 结构 (行号: 139-170)
```cpp
struct TypeStruct {
    Slice<Entity *> fields;
    String *tags;
    i64 *offsets;
    Ast *node;
    Scope *scope;
    i64 custom_align;
    i64 custom_min_field_align;
    i64 custom_max_field_align;
    Type *polymorphic_params;
    Type *polymorphic_parent;
    Wait_Signal polymorphic_wait_signal;
    Type *soa_elem;
    i32 soa_count;
    StructSoaKind soa_kind;
    bool is_polymorphic;
    bool are_offsets_set : 1;
    bool is_packed : 1;
    bool is_raw_union : 1;
    bool is_all_or_none : 1;
    bool is_simple : 1;
    bool is_poly_specialized : 1;
};
```
```go
// 建议Go实现
type TypeStruct struct {
    Fields      []Entity
    Tags        []string
    Offsets     []int64
    Node        *Ast
    Scope       *Scope
    CustomAlign int64
    // ...其他字段
    IsPolymorphic bool
    IsPacked    bool
}
```

#### TypeProc 结构 (行号: 190-212)
```cpp
struct TypeProc {
    Ast *node;
    Scope *scope;
    Type *params;   // Type_Tuple
    Type *results;  // Type_Tuple
    i32 param_count;
    i32 result_count;
    isize specialization_count;
    ProcCallingConvention calling_convention;
    i32 variadic_index;
    String require_target_feature;
    String enable_target_feature;
    bool variadic;
    bool require_results;
    bool c_vararg;
    bool is_polymorphic;
    bool is_poly_specialized;
    bool has_named_results;
    bool diverging;
    bool return_by_pointer;
    bool optional_ok;
};
```
```go
// 建议Go实现 - 函数签名
type TypeProc struct {
    Node        *Ast
    Scope      *Scope
    Params     *TypeTuple  // 参数类型列表
    Results    *TypeTuple  // 返回类型列表
    Variadic   bool
    Diverging  bool  // 无返回值
    // ...
}
```

#### TypeUnion 结构 (行号: 172-188)
```cpp
struct TypeUnion {
    Slice<Type *> variants;
    Ast *node;
    Scope *scope;
    std::atomic<i64> variant_block_size;
    i64 custom_align;
    Type *polymorphic_params;
    Type *polymorphic_parent;
    Wait_Signal polymorphic_wait_signal;
    std::atomic<i16> tag_size;
    bool is_polymorphic;
    bool is_poly_specialized;
    UnionTypeKind kind;
};
```

### Selection 结构 (行号: 434-442)
```cpp
struct Selection {
    Entity *entity;
    Array<i32> index;
    bool indirect;
    u8 swizzle_count;
    u8 swizzle_indices;
    bool is_bit_field;
    bool pseudo_field;
};
```

## 基础类型表 (行号: 483-5603)
```cpp
gb_global Type basic_types[] = {
    {Type_Basic, {Basic_Invalid, 0, 0, STR_LIT("invalid type")}},
    {Type_Basic, {Basic_llvm_bool, BasicFlag_Boolean|BasicFlag_LLVM, 1, STR_LIT("llvm bool")}},
    {Type_Basic, {Basic_bool, BasicFlag_Boolean, 1, STR_LIT("bool")}},
    {Type_Basic, {Basic_i8, BasicFlag_Integer, 1, STR_LIT("i8")}},
    // ... 所有基础类型定义
};
```
```go
// 建议Go实现
var BasicTypes = map[BasicKind]*Type{
    BasicInvalid: {Kind: TypeBasic, Basic: &BasicType{Kind: BasicInvalid}},
    BasicBool: {Kind: TypeBasic, Basic: &BasicType{Kind: BasicBool, Size: 1}},
    BasicI8: {Kind: TypeBasic, Basic: &BasicType{Kind: BasicI8, Size: 1}},
    BasicU8: {Kind: TypeBasic, Basic: &BasicType{Kind: BasicU8, Size: 1}},
    // ...
}
```

## 函数签名

### 类型检查函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `is_type_comparable(t)` | 类型是否可比较 | 简单 | 411 |
| `is_type_simple_compare(t)` | 是否简单可比较 | 简单 | 412 |
| `type_deref(t, allow_multi)` | 解引用指针类型 | 内部方法 | 413 |
| `base_type(t)` | 获取基础类型 | 内部方法 | 414 |
| `alloc_type_multi_pointer(e)` | 创建多指针 | 内部方法 | 415 |
| `type_info_flags_of_type(t)` | 获取类型信息标志 | 简单 | 417 |

### Selection操作
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `make_selection(e, index, indirect)` | 创建选择器 | 工厂方法 | 445 |
| `selection_add_index(s, index)` | 添加索引 | 内部方法 | 450 |
| `selection_combine(lhs, rhs)` | 组合选择器 | 合并逻辑 | 457 |
| `sub_selection(sel, offset)` | 子选择器 | 偏移 | 466 |
| `trim_selection(sel)` | 修剪选择器 | 去掉最后索引 | 474 |

## 重写建议

### 核心重写要点
1. **Type联合体** - 在Go中使用接口或类型断言
2. **基础类型表** - 使用map替代数组便于查询
3. **Atomic操作** - 使用Go的atomic包

### 可用Go标准库
- **反射**: `reflect` 包可处理很多类型信息
- **原子操作**: `sync/atomic` 包

### 需要自行实现
1. **类型缓存**: 使用map缓存创建的Type
2. **泛型参数**: Go 1.18+泛型可简化实现
3. **SOA支持**: 结构数组布局需要自行实现

### 重写优先级
| 优先级 | 模块 | 重写必要性 |
|--------|------|-----------|
| 高 | Type定�� | 需要适配 |
| 高 | 基础类型表 | 需要重构为map |
| 中 | TypeStruct/Proc/Union | 使用Go结构体 |
| 低 | 原子缓存 | 使用atomic |

## 总结
- **整体重写必要性**: 中等偏高
- **建议策略**: 
  - 基础类型用map管理
  - 复合类型用Go结构体重写
  - 利用Go反射简化类型操作
  - 泛型(Go 1.18+)可很好支持多态类型