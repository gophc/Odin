# checker.cpp 定义文档

## 文件信息
- **源文件**: src/checker.cpp
- **核心功能**: 语义分析器，执行类型检查、作用域管理、符号解析

## 类型定义

### Scope 作用域
```cpp
struct Scope {
    Scope *parent;
    AstPackage *pkg;
    Map<InternedString, Entity *> elements;
    Map<InternedString, Scope *> scopes;
    // ...
};
```
```go
// 建议Go实现
type Scope struct {
    Parent  *Scope
    Package *AstPackage
    Elements map[string]*Entity  // 符号表
    Children map[string]*Scope // 子作用域
}
```

### Entity 实体
```cpp
enum EntityKind {
    Entity_Invalid,
    Entity_Constant,
    Entity_Variable,
    Entity_TypeName,
    Entity_Procedure,
    Entity_ProcGroup,
    Entity_Builtin,
    Entity_ImportName,
    Entity_LibraryName,
    Entity_Nil,
    Entity_Label,
};

struct Entity {
    EntityKind kind;
    Token token;        // 定义位置的token
    TokenPos pos;
    DeclInfo decl;
    // 根据kind有不同的联合...
};
```

### DeclInfo 声明信息
```cpp
struct DeclInfo {
    Scope *scope;
    DeclInfo *parent;
    Ast *file;
    Array<CommentGroup *> docs;
    // ...
};
```

### CheckerContext 检查器上下文
```cpp
struct CheckerContext {
    CheckerInfo *info;
    Scope *scope;
    Array<CheckerContext> prev_scope_stack;
    // ...
};
```

## 函数签名

### 作用域管理
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `create_scope(info, parent)` | 创建作用域 | new(Scope) | - |
| `create_scope_from_file(info, f)` | 从文件创建 | new(Scope) | - |
| `create_scope_from_package(c, pkg)` | 从包创建 | new(Scope) | - |
| `destroy_scope(scope)` | 销毁作用域 | GC处理 | - |
| `add_scope(c, node, scope)` | 添加作用域 | map设置 | - |
| `scope_of_node(node)` | 获取节点作用域 | map查询 | - |
| `check_open_scope(c, node)` | 打开作用域 | 压栈 | - |
| `check_close_scope(c, node)` | 关闭作用域 | 出栈 | - |

### 符号查询
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `scope_lookup_current(s, name, hash)` | 当前作用域查询 | map查找 | - |
| `scope_lookup_parent(s, name, ...)` | 父作用域查询 | 递归查找 | - |
| `scope_insert(s, entity)` | 插入符号 | map设置 | - |

### 声明处理
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `init_decl_info(d, scope, parent)` | 初始化声明信息 | - |
| `make_decl_info(scope, parent)` | 创建声明信息 | - |
| `add_dependency(info, d, e)` | 添加依赖 | - |

### 检查功能
| 函数名 | 功能 | 行号 |
|--------|------|------|
| `check_vet_flags(c)` | vet标志检查 | - |
| `check_scope_usage(c, scope, flags)` | 作用域使用检查 | - |
| `check_vet_shadowing(c, entity, ve)` | 遮蔽检查 | - |
| `check_vet_unused(c, entity, ve)` | 未使用检查 | - |

## 重写建议

### 核心重写要点
1. **作用域** - 直接使用Go的闭包和map
2. **Entity** - 使用map[string]Entity
3. **检查上下文** - 使用栈结构

### 可用Go实现
- **符号解析**: 使用map[string]*Entity
- **作用域链**: 使用嵌套结构体+闭包
- **类型检查**: 使用反射或接口

### 重写结构
```go
type Checker struct {
    Scopes    []*Scope
    Current  *Scope
    Entities map[string]*Entity // 全局Entity表
}

func (c *Checker) EnterScope() {
    c.Scopes = append(c.Scopes, &Scope{Parent: c.Current})
    c.Current = c.Scopes[len(c.Scopes)-1]
}

func (c *Checker) ExitScope() {
    c.Current = c.Current.Parent
}

func (c *Checker) Lookup(name string) *Entity {
    for s := c.Current; s != nil; s = s.Parent {
        if e, ok := s.Elements[name]; ok {
            return e
        }
    }
    return nil
}
```

## 总结
- **整体重写必要性**: 中等
- **建议策略**: 
  - 作用域用Go结构体实现
  - Entity用map管理
  - 栈结构管理作用域嵌套