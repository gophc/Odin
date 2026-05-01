# tokenizer.cpp 定义文档

## 文件信息
- **源文件**: src/tokenizer.cpp
- **行数范围**: 1-1128
- **核心功能**: 词法分析器，将源代码转换为Token序列

## 类型定义

### TokenKind 枚举
```cpp
// 核心Token类型 - 行号: 128-132
enum TokenKind : u8 {
    Token_Invalid,
    Token_EOF,
    Token_Comment,
    Token_FileTag,
    Token_Ident,        // 标识符
    Token_Integer,     // 整数
    Token_Float,      // 浮点数
    Token_String,     // 字符串
    // 运算符...
    Token_Eq, Token_Not, Token_Add, Token_Sub, Token_Mul, Token_Quo, ...
    // 关键字...
    Token_import, Token_package, Token_proc, Token_struct, Token_union, ...
    // 结束
    Token_Count,
};
```
```go
// 建议Go实现
type TokenKind uint8

const (
    TokenInvalid TokenKind = iota
    TokenEOF
    TokenComment
    TokenFileTag
    TokenIdent
    TokenInteger
    TokenFloat
    TokenString
    // ...其他TokenKind
)
```

### Token 结构
```cpp
struct Token {
    TokenKind kind;
    u8        flags;
    String    string;
    TokenPos  pos;
};
```
```go
// 建议Go实现
type Token struct {
    Kind   TokenKind
    Flags uint8
    Text  String  // 嵌入String而不是string便于复用
    Pos   TokenPos
}

type TokenPos struct {
    FileID int32  // 文件ID
    Offset int32  // 字符偏移
    Line   int32  // 行号(从1开始)
    Column int32  // 列号(从1开始)
}
```

### Tokenizer 结构
```cpp
struct Tokenizer {
    i32 curr_file_id;
    String fullpath;
    u8 *start;
    u8 *end;
    Rune  curr_rune;
    u8 *  curr;
    u8 *  read_curr;
    i32   column_minus_one;
    i32   line_count;
    i32   error_count;
    bool insert_semicolon;
    LoadedFile loaded_file;
};
```
```go
// 建议Go实现 - 行号: 297-314
type Tokenizer struct {
    FileID       int32
    FullPath     String
    Data        []byte // start-end 合并为Data
    Pos          int   // 当前位置索引
    ReadPos     int   // 预读位置
    CurrRune    rune
    Column      int32 // column_minus_one + 1
    Line        int32
    ErrorCount  int32
    InsertSemi bool
    LoadedFile LoadedFile
}
```

## 函数签名

### 初始化函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `init_keyword_hash_table()` | 初始化关键字哈希表 | 构建map | 176 |
| `init_tokenizer_with_data(t, path, data, size)` | 用数据初始化 | 直接赋值 | 380 |
| `init_tokenizer_from_fullpath(t, path, copy)` | 从文件加载 | 读取文件 | 404 |

### 词法扫描函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `advance_to_next_rune(t)` | 读取下一个字符 | 内部方法 | 350 |
| `digit_value(r)` | 数字字符值 | switch实现 | 425 |
| `scan_mantissa(t, base, force)` | 扫描数字尾随部分 | 扫描逻辑 | 437 |
| `peek_byte(t, offset)` | 查看字节 | Data[pos+offset] | 446 |
| `scan_number_to_token(t, token, seen_point)` | 扫描数字字面量 | 核心解析 | 453 |

### Token判定函数
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `token_is_literal(t)` | 是否字面量 | 范围检查 | 264 |
| `token_is_operator(t)` | 是否运算符 | 范围检查 | 267 |
| `token_is_keyword(t)` | 是否关键字 | 范围检查 | 270 |
| `token_is_comparison(t)` | 是否比较符 | 范围检查 | 273 |
| `token_is_shift(t)` | 是否位移符 | 具体检查 | 276 |

### TokenPos操作
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `token_pos_cmp(a, b)` | 比较位置 | 比较逻辑 | 209 |
| `token_pos_add_column(pos)` | 列前进 | 内部方法 | 230 |

## 关键字哈希表 - 需要Go实现
```cpp
// 行号: 147-156
enum { 
    KEYWORD_HASH_TABLE_COUNT = 1<<9,
    KEYWORD_HASH_TABLE_MASK = 511,
};
gb_global KeywordHashEntry keyword_hash_table[512];
gb_global isize max_keyword_size = 11;
```
```go
// 建议Go实现
var keywordHashTable [512]KeywordEntry
var maxKeywordSize int = 11
type KeywordEntry struct {
    Hash uint32
    Kind TokenKind
    Text String
}
```

## 重写建议

### 核心重写要点
1. **Tokenizer** - 需要重写，Go的rune处理不同
2. **关键字表** - 使用map[string]TokenKind替代哈希表
3. **位置追踪** - 需要维护Line/Column

### 可用Go标准库
- **字符分类**: `unicode` 包
- **UTF8解码**: `unicode/utf8` 包
- **数字解析**: `strconv` 包

### 需要自行实现
1. **Token生成**: 分析每个Token类型对应正则
2. **关键字识别**: 使用map替代哈希表
3. **预读机制**: 实现peek功能

### 重写优先级
| 优先级 | 模块 | 重写必要性 |
|--------|------|-----------|
| 高 | Tokenizer解析逻辑 | 需要完全重写 |
| 高 | Token定义 | 需要适配 |
| 中 | 关键字表 | 可简化 |
| 低 | 错误处理 | 可简化 |

## 总结
- **整体重写必要性**: 高
- **建议策略**: 
  - Token和TokenPos结构需适配Go
  - 分词逻辑需要重写
  - 关键字表使用map简化
  - 可利用unicode包辅助字符分类