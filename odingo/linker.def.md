# linker.cpp 定义文档

## 文件信息
- **源文件**: src/linker.cpp
- **行数范围**: 1-1025
- **核心功能**: 链接器，调用系统链接器生成最终可执行文件

## 类型定义

### LinkerData 链接器数据
```cpp
struct LinkerData {
    BlockingMutex foreign_mutex;
    PtrSet<Entity *> foreign_libraries_set;
    Array<Entity *> foreign_libraries;
    Array<String> output_object_paths;
    Array<String> output_temp_paths;
    String output_base;
    String output_name;
    bool needs_system_library_linked;
};
```
```go
// 建议Go实现
type LinkerData struct {
    mu             sync.Mutex
    ForeignLibs   map[*Entity]bool
    ObjectPaths   []string
    TempPaths     []string
    OutputBase    string
    OutputName   string
    NeedSysLib   bool
}
```

## 函数签名

### 初始化
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `linker_data_init(ld, info, path)` | 初始化链接数据 | new(LinkerData) | 20 |
| `linker_enable_system_library_linking(ld)` | 启用系统库链接 | 设置标志 | 16 |

### 链接阶段
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `linker_stage(gen)` | 执行链接 | 调用外部链接器 | 55 |

### 系统调用
| 函数名 | 功能 | Go建议 | 行号 |
|--------|------|--------|------|
| `system_exec_command_line_app(name, fmt, ...)` | 执行命令行 | os/exec | - |
| `system_exec_command_line_app_output(cmd, output)` | 执行并捕获输出 | os/exec.Output | - |

## 链接器选择
```cpp
enum LinkerChoice {
    Linker_Default,
    Linker_lld,
    Linker_mold,
    Linker_radlink,
};
```

### 支持的平台
- Windows: MSVC link.exe, lld-link, rad-link
- macOS: ld64
- Linux: ld, lld, mold

## 重写建议

### 核心重写要点
1. **外部调用** - 使用 `os/exec` 包调用系统链接器
2. **参数构造** - 使用 `flag` 或自定义参数构建
3. **库管理** - 使用map管理外部库

### 可用Go标准库
- **执行命令**: `os/exec`
- **路径处理**: `path/filepath`
- **标志解析**: `flag` 或 `pflag`

### 重写结构
```go
type Linker struct {
    Data LinkerData
    Cmd  *exec.Cmd
}

func (l *Linker) Link() error {
    // 根据平台选择链接器
    cmd := exec.Command("link.exe", l.Args()...)
    return cmd.Run()
}
```

## 总结
- **整体重写必要性**: 中等
- **建议策略**: 
  - 使用os/exec调用系统链接器
  - 参数构造逻辑可复用
  - 平台检测使用goarch/goos