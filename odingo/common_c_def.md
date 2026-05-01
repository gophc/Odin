# common.i.cpp 函数和类型定义文档

本文档整理 `src/cipp/common.i.cpp` 中的所有导出函数和类型，按源文件分区组织。

---

## 1. src/common.cpp (行 1-23)

### 函数声明

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5 | `static gbAllocator heap_allocator(void)` | 获取堆分配器 |
| 6 | `static i32 next_pow2(i32 n)` | 计算大于等于n的最小2的幂 |
| 7 | `static i64 next_pow2(i64 n)` | 64位版本 |
| 8 | `static isize next_pow2_isize(isize n)` | isize版本 |
| 9 | `static void debugf(char const *fmt, ...)` | 调试打印 |

### 类型声明

| 行号 | 类型 | 说明 |
|------|------|------|
| 11-13 | `TypeIsPointer<T>` | 模板: 判断类型是否指针 |
| 15-17 | `TypeIsPointer<T *>` | 模板特化: 指针版本返回true |
| 18-20 | `TypeIsPtrSizedInteger<T>` | 判断是否是isize/usize |
| 21-23 | `TypeIs64BitInteger<T>` | 判断是否是i64/u64 |

---

## 2. src/array.cpp (行 25-422)

### 类型声明

| 行号 | 类型 | 说明 |
|------|------|------|
| 28-41 | `Array<T>` | 动态数组结构体，含allocator、data、count、capacity |
| 70-81 | `Slice<T>` | 切片结构体，含data、count |
| 26 | `static_assertion` | 静态断言确保sizeof(T)正确 |

### Array 函数声明

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 207 | `void array_init(Array<T> *, gbAllocator const &)` | 初始化数组，容量为8 |
| 212 | `void array_init(Array<T> *, gbAllocator const &, isize count)` | 初始化指定数量 |
| 216 | `void array_init(Array<T> *, gbAllocator const &, isize count, isize capacity)` | 初始化指定数量和容量 |
| 226 | `Array<T> array_make_from_ptr(T *data, isize count, isize capacity)` | 从已有指针创建数组 |
| 234 | `Array<T> array_make(gbAllocator const &a)` | 创建空数组 |
| 244 | `Array<T> array_make(gbAllocator const &a, isize count)` | 创建指定数量数组 |
| 253 | `Array<T> array_make(gbAllocator const &a, isize count, isize capacity)` | 创建指定容量数组 |
| 262 | `void array_free(Array<T> *array)` | 释放数组内存 |
| 270 | `void array__grow(Array<T> *array, isize min_capacity)` | 内部: 扩容 |
| 278 | `void array_add(Array<T> *array, T const &t)` | 添加元素 |
| 293 | `T *array_add_and_get(Array<T> *array)` | 添加并返回指针 |
| 303 | `void array_add_elems(Array<T> *array, T const *elems, isize elem_count)` | 添加多个元素 |
| 312 | `T array_pop(Array<T> *array)` | 弹出最后一个元素 |
| 318 | `void array_clear(Array<T> *array)` | 清空数组 |
| 322 | `void array_reserve(Array<T> *array, isize capacity)` | 预留容量 |
| 328 | `void array_resize(Array<T> *array, isize count)` | 调整大小 |
| 335 | `void array_set_capacity(Array<T> *array, isize capacity)` | 设置容量 |
| 358 | `Array<T> array_slice(Array<T> const &array, isize lo, isize hi)` | 切片操作 |
| 370 | `Array<T> array_clone(gbAllocator const &allocator, Array<T> const &array)` | 克隆数组 |
| 376 | `void array_ordered_remove(Array<T> *array, isize index)` | 有序删除元素 |
| 383 | `void array_unordered_remove(Array<T> *array, isize index)` | 无序删除(交换最后元素) |
| 192 | `void array_copy(Array<T> *array, Array<T> const &data, isize offset)` | 复制数据 |
| 196 | `void array_copy(Array<T> *array, Array<T> const &data, isize offset, isize count)` | 复制指定数量 |
| 200 | `T *array_end_ptr(Array<T> *array)` | 获取最后一个元素的指针 |
| 66 | `void array_sort(Array<T> &array, gbCompareProc compare_proc)` | 排序 |

### Slice 函数声明

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 114 | `Slice<T> slice_from_array(Array<T> const &a)` | 从Array创建Slice |
| 85 | `Slice<T> slice_make(gbAllocator const &allocator, isize count)` | 创建Slice |
| 96 | `void slice_init(Slice<T> *s, gbAllocator const &allocator, isize count)` | 初始化Slice |
| 105 | `void slice_free(Slice<T> *s, gbAllocator const &allocator)` | 释放Slice |
| 109 | `void slice_resize(Slice<T> *s, gbAllocator const &allocator, isize new_count)` | 调整Slice大小 |
| 118 | `Slice<T> slice_array(Array<T> const &array, isize lo, isize hi)` | Array切片 |
| 129 | `Slice<T> slice_clone(gbAllocator const &allocator, Slice<T> const &a)` | 克隆Slice |
| 134 | `Slice<T> slice_clone_from_array(gbAllocator const &allocator, Array<T> const &a)` | 从Array克隆 |
| 139 | `void slice_copy(Slice<T> *slice, Slice<T> const &data)` | 复制 |
| 144 | `void slice_copy(Slice<T> *slice, Slice<T> const &data, isize offset)` | 偏移复制 |
| 149 | `void slice_copy(Slice<T> *slice, Slice<T> const &data, isize offset, isize count)` | 偏移限量复制 |
| 154 | `Slice<T> slice(Slice<T> const &array, isize lo, isize hi)` | Slice切片 |
| 165 | `Slice<T> slice(Array<T> const &array, isize lo, isize hi)` | Array切片(重载) |
| 176 | `void slice_ordered_remove(Slice<T> *array, isize index)` | 有序删除 |
| 183 | `void slice_unordered_remove(Slice<T> *array, isize index)` | 无序删除 |

### 迭代器函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 392 | `T *begin(Array<T> &array)` | Array起始迭代器 |
| 396 | `T const *begin(Array<T> const &array)` | Array起始迭代器(const) |
| 400 | `T *end(Array<T> &array)` | Array结束迭代器 |
| 404 | `T const *end(Array<T> const &array)` | Array结束迭代器(const) |
| 408 | `T *begin(Slice<T> &array)` | Slice起始迭代器 |
| 412 | `T const *begin(Slice<T> const &array)` | Slice起始迭代器(const) |
| 416 | `T *end(Slice<T> &array)` | Slice结束迭代器 |
| 420 | `T const *end(Slice<T> const &array)` | Slice结束迭代器(const) |

---

## 3. src/threading.cpp (行 423-786)

### 类型声明

| 行号 | 类型 | 说明 |
|------|------|------|
| 426 | `BlockingMutex` | 阻塞互斥锁 |
| 427 | `RecursiveMutex` | 递归互斥锁 |
| 428 | `RwMutex` | 读写锁 |
| 429 | `Semaphore` | 信号量 |
| 430 | `Condition` | 条件变量 |
| 431 | `Thread` | 线程结构 |
| 432 | `ThreadPool` | 线程池 |
| 433 | `Parker` | 线程暂停/唤醒 |
| 436-439 | `WorkerTask` | 工作任务结构 |
| 440-443 | `TaskRingBuffer` | 环形缓冲区 |
| 444-448 | `TaskQueue` | 任务队列 |
| 449-457 | `Thread` | 线程详细信息 |
| 458 | `Futex` | 快速用户区锁(原子int32) |
| 459 | `Footex` | 易失性int32 |
| 490 | `Wait_Signal` | 等待信号 |

### 线程函数声明

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 434 | `isize thread_pool_thread_proc(Thread *thread)` | 线程池工作线程处理函数 |
| 460 | `void futex_wait(Futex *addr, Footex val)` | 等待futex |
| 461 | `void futex_signal(Futex *addr)` | 信号futex |
| 462 | `void futex_broadcast(Futex *addr)` | 广播futex |
| 463 | `void mutex_lock(BlockingMutex *m)` | 加锁(阻塞) |
| 464 | `bool mutex_try_lock(BlockingMutex *m)` | 尝试加锁 |
| 465 | `void mutex_unlock(BlockingMutex *m)` | 解锁 |
| 466 | `void mutex_lock(RecursiveMutex *m)` | 加锁(递归) |
| 467 | `bool mutex_try_lock(RecursiveMutex *m)` | 尝试加锁(递归) |
| 468 | `void mutex_unlock(RecursiveMutex *m)` | 解锁(递归) |
| 469 | `void rw_mutex_lock(RwMutex *m)` | 写锁 |
| 470 | `bool rw_mutex_try_lock(RwMutex *m)` | 尝试写锁 |
| 471 | `void rw_mutex_unlock(RwMutex *m)` | 解写锁 |
| 472 | `void rw_mutex_shared_lock(RwMutex *m)` | 读锁 |
| 473 | `bool rw_mutex_try_shared_lock(RwMutex *m)` | 尝试读锁 |
| 474 | `void rw_mutex_shared_unlock(RwMutex *m)` | 解读锁 |
| 475 | `void semaphore_post(Semaphore *s, i32 count)` | 信号量增加 |
| 476 | `void semaphore_wait(Semaphore *s)` | 信号量等待 |
| 477 | `void condition_broadcast(Condition *c)` | 条件广播 |
| 478 | `void condition_signal(Condition *c)` | 条件信号 |
| 479 | `void condition_wait(Condition *c, BlockingMutex *m)` | 条件等待 |
| 480 | `void park(Parker *p)` | 暂停线程 |
| 481 | `void unpark_one(Parker *p)` | 唤醒一个线程 |
| 482 | `void unpark_all(Parker *p)` | 唤醒所有线程 |
| 483 | `u32 thread_current_id(void)` | 获取当前线程ID |
| 484 | `void thread_init(ThreadPool *pool, Thread *t, isize idx)` | 初始化线程 |
| 485 | `void thread_init_and_start(ThreadPool *pool, Thread *t, isize idx)` | 初始化并启动线程 |
| 486 | `void thread_join_and_destroy(Thread *t)` | 等待并销毁线程 |
| 487 | `void thread_set_name(Thread *t, char const *name)` | 设置线程名 |
| 488 | `void yield_thread(void)` | 让出CPU |
| 489 | `void yield_process(void)` | 让出进程 |

### WaitSignal 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 493 | `void wait_signal_until_available(Wait_Signal *ws)` | 等待信号可用 |
| 498 | `void wait_signal_set(Wait_Signal *ws)` | 设置信号 |

### 锁常量

| 行号 | 常量 | 值 |
|------|------|-----|
| 647 | `RWLOCK_WRITER` | 1<<0 |
| 648 | `RWLOCK_UPGRADED` | 1<<1 |
| 649 | `RWLOCK_READER` | 1<<2 |

### RWSpinLock 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 653 | `void rwlock_release_write(RWSpinLock *l)` | 释放写锁 |
| 657 | `bool rwlock_try_acquire_upgrade(RWSpinLock *l)` | 尝试升级 |
| 661 | `void rwlock_acquire_upgrade(RWSpinLock *l)` | 获取升级 |
| 666 | `void rwlock_release_upgrade(RWSpinLock *l)` | 释放升级 |
| 670 | `bool rwlock_try_release_upgrade_and_acquire_write(RWSpinLock *l)` | 尝试升级并获取写锁 |
| 674 | `void rwlock_release_upgrade_and_acquire_write(RWSpinLock *l)` | 升级并获取写锁 |

### Parker 状态枚举

| 行号 | 枚举值 | 说明 |
|------|--------|------|
| 682 | `ParkerState_Empty` | 0 |
| 683 | `ParkerState_Notified` | 1 |
| 684 | `ParkerState_Parked` | 0xffffffffui32 |

### MutexGuard (RAII锁)

| 行号 | 类型 | 说明 |
|------|------|------|
| 502-537 | `MutexGuard` | RAII互斥锁守卫 |

### 线程内部函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 709 | `u32 thread_current_id(void)` | 获取线程ID |
| 714 | `void yield_thread(void)` | CPU让出 |
| 717 | `void yield(void)` | 让出 |
| 720 | `DWORD __stdcall internal_thread_proc(void *arg)` | 内部线程过程 |
| 725 | `TaskRingBuffer *task_ring_init(isize size)` | 初始化环形缓冲 |
| 731 | `void thread_queue_destroy(TaskQueue *q)` | 销毁队列 |

---

## 4. src/common_memory.cpp (行 817-1233)

### 内存函数声明

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 818-821 | `U bit_cast(V &v)` | 位转换模板 |
| 822-825 | `i64 align_formula(i64 size, i64 align)` | 对齐公式(i64) |
| 826-829 | `isize align_formula_isize(isize size, isize align)` | 对齐公式(isize) |
| 830-833 | `void *align_formula_ptr(void *ptr, isize align)` | 指针对齐公式 |
| 836 | `void virtual_memory_init(void)` | 虚拟内存初始化 |
| 860 | `isize arena_align_forward_offset(Arena *arena, isize alignment)` | Arena对齐偏移 |

### 内存块结构

| 行号 | 类型 | 说明 |
|------|------|------|
| 840-846 | `MemoryBlock` | 内存块结构(prev, base, size, used, committed) |
| 847-853 | `Arena` | Arena内存分配器(curr_block, minimum_block_size, temp_count, parent_thread, custom_arena) |

### 常量

| 行号 | 常量 | 值 |
|------|------|-----|
| 854 | `DEFAULT_MINIMUM_BLOCK_SIZE` | 8ll*1024ll*1024ll (8MB) |
| 855 | `DEFAULT_PAGE_SIZE` | 4096 |
| 909 | `STATIC_ARENA_DEFAULT_COMMIT_BLOCK_SIZE` | 8<<20 |

### StaticArena 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 902-908 | `StaticArena` | 静态Arena(data, used, committed, reserved, commit_block_size) |

### Arena 函数声明

| 行号 | ���数���名 | 说明 |
|------|----------|------|
| 869 | `void thread_init_arenas(Thread *t)` | 初始化线程Arena |
| 877 | `void *arena_alloc(Arena *arena, isize min_size, isize alignment)` | Arena分配 |
| 886 | `void arena_free_all(Arena *arena)` | 释放所有Arena内存 |
| 910 | `void static_arena_init(StaticArena *arena, isize reserve_size, isize commit_block_size)` | 初始化静态Arena |
| 918 | `void static_arena_commit_memory(StaticArena *arena, isize amount)` | 提交静态内存 |
| 927 | `void *static_arena_alloc(StaticArena *arena, isize size, isize alignment)` | 静态Arena分配 |
| 942 | `T *arena_alloc_item(Arena *arena)` | 分配单个item |

### PlatformMemoryBlock

| 行号 | 类型 | 说明 |
|------|------|------|
| 952-956 | `PlatformMemoryBlock` | 平台内存块 |

### ArenaTemp

| 行号 | 类型 | 说明 |
|------|------|------|
| 1048-1052 | `ArenaTemp` | Arena临时状态 |

### ArenaTemp 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1053 | `ArenaTemp arena_temp_begin(Arena *arena)` | 开始临时状态 |
| 1065 | `void arena_temp_end(ArenaTemp const &temp)` | 结束临时状态 |
| 1096 | `void arena_temp_ignore(ArenaTemp const &temp)` | 忽略临时状态 |

### ArenaTempGuard

| 行号 | 类型 | 说明 |
|------|------|------|
| 1103-1111 | `ArenaTempGuard` | RAII临时状态守卫 |

### Arena 分配器

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1112 | `void *arena_allocator_proc(...)` | Arena分配器回调 |
| 1113 | `gbAllocator arena_allocator(Arena *arena)` | 获取Arena分配器 |

### ThreadArenaKind

| 行号 | 类型 | 说明 |
|------|------|------|
| 1146-1149 | `ThreadArenaKind` | 线程Arena类型(永久/临时) |

### 默认Arena

| 行号 | 变量 | 说明 |
|------|------|------|
| 1150 | `default_permanent_arena` | 默认永久Arena |
| 1151 | `default_temporary_arena` | 默认临时Arena |

### 获取Arena

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1152 | `Arena *get_arena(ThreadArenaKind kind)` | 获取线程Arena |

### 分配模板函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1161 | `T *arena_alloc_array(Arena *arena, isize count)` | 分配数组 |
| 1166 | `T *permanent_alloc_item()` | 永久分配单项 |
| 1171 | `T *permanent_alloc_array(isize count)` | 永久分配数组 |
| 1176 | `Slice<T> permanent_slice_make(isize count)` | 永久创建Slice |
| 1182 | `T *temporary_alloc_item()` | 临时分配单项 |
| 1187 | `T *temporary_alloc_array(isize count)` | 临时分配数组 |
| 1192 | `Slice<T> temporary_slice_make(isize count)` | 临时创建Slice |

### 分配器获取

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1197 | `void *thread_arena_allocator_proc(...)` | 线程Arena回调 |
| 1224 | `gbAllocator permanent_allocator()` | 永久分配器 |
| 1227 | `gbAllocator temporary_allocator()` | 临时分配器(同永久) |

### Heap分配器

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1230 | `bool IS_ODIN_DEBUG(void)` | 是否调试模式 |
| 1232 | `gbAllocator heap_allocator(void)` | 获取堆分配器 |
| 1239 | `void *heap_allocator_proc(...)` | 堆分配器回调 |

### 数组Raw调整

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1274 | `isize resize_array_raw(T **array, ...)` | 调整原始数组 |

---

## 5. src/queue.cpp (行 1293-1484)

### MPSCQueue 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 1295-1297 | `MPSCNode<T>` | 单生产者单消费者节点 |
| 1300-1305 | `MPSCQueue<T>` | 单生产者单消费者队列 |

### MPSCQueue 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1311 | `void mpsc_init(MPSCQueue<T> *q, gbAllocator const &allocator)` | 初始化 |
| 1318 | `void mpsc_destroy(MPSCQueue<T> *q)` | 销毁 |
| 1322 | `MPSCNode<T> *mpsc_alloc_node(MPSCQueue<T> *q, T const &value)` | 分配节点 |
| 1328 | `void mpsc_free_node(MPSCQueue<T> *q, MPSCNode<T> *node)` | 释放节点 |
| 1331 | `isize mpsc_enqueue(MPSCQueue<T> *q, MPSCNode<T> *node)` | 入队 |
| 1339 | `isize mpsc_enqueue(MPSCQueue<T> *q, T const &value)` | 入队(值) |
| 1344 | `bool mpsc_dequeue(MPSCQueue<T> *q, T *value_)` | 出队 |

### MPMCQueue 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 1357 | `MPMCQueueAtomicIdx` | 原子索引类型 |
| 1359-1370 | `MPMCQueue<T>` | 多生产者多消费者队列 |

### MPMCQueue 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1371 | `gbAllocator mpmc_allocator(void)` | 获取分配器 |
| 1374 | `void mpmc_internal_init_indices(MPMCQueueAtomicIdx *indices, i32 offset, i32 size)` | 初始化索引 |
| 1390 | `void mpmc_init(MPMCQueue<T> *q, isize size_i)` | 初始化 |
| 1405 | `void mpmc_destroy(MPMCQueue<T> *q)` | 销毁 |
| 1411 | `bool mpmc_internal_grow(MPMCQueue<T> *q)` | 内部扩容 |
| 1434 | `i32 mpmc_enqueue(MPMCQueue<T> *q, T const &data)` | 入队 |
| 1460 | `bool mpmc_dequeue(MPMCQueue<T> *q, T *data_)` | 出队 |

---

## 6. src/string.cpp (行 1485-2058)

### String 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 1487-1493 | `String` | UTF-8字符串(text, len) |
| 1495-1497 | `String_Iterator` | 字符串迭代器 |
| 1499-1505 | `String16` | UTF-16字符串 |

### 字符串创建函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1507 | `String make_string(u8 const *text, isize len)` | 创建字符串 |
| 1516 | `String16 make_string16(u16 const *text, isize len)` | 创建String16 |
| 1522 | `isize string16_len(u16 const *s)` | String16长度 |
| 1532 | `String make_string_c(char const *text)` | 从C字符串创建 |
| 1535 | `String16 make_string16_c(u16 const *text)` | 从C字符串创建String16 |

### 字符串操作函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1538 | `String substring(String const &s, isize lo, isize hi)` | 子字符串 |
| 1543 | `String16 substring(String16 const &s, isize lo, isize hi)` | String16子串 |
| 1548 | `char *alloc_cstring(gbAllocator a, String s)` | 分配C字符串 |
| 1554 | `wchar_t *alloc_wstring(gbAllocator a, String16 s)` | 分配宽字符字符串 |
| 1560 | `bool str_eq_ignore_case(String const &a, String const &b)` | 忽略大小写比较 |
| 1573 | `bool str_eq_ignore_case(String const &a, char const (&b)[N])` | 忽略大小写比较(C字符串) |
| 1581 | `void string_to_lower(String *s)` | 转小写 |
| 1586 | `int string_compare(String const &a, String const &b)` | 比较字符串 |
| 1603 | `int string16_compare(String16 const &a, String16 const &b)` | 比较String16 |
| 1620 | `isize string_index_byte(String const &s, u8 x)` | 查找字节位置 |
| 1628 | `bool str_eq(String const &a, String const &b)` | 字符串相等 |
| 1634 | `bool str_ne(String const &a, String const &b)` | 字符串不等 |
| 1635 | `bool str_lt(String const &a, String const &b)` | 小于 |
| 1636 | `bool str_gt(String const &a, String const &b)` | 大于 |
| 1637 | `bool str_le(String const &a, String const &b)` | 小于等于 |
| 1638 | `bool str_ge(String const &a, String const &b)` | 大于等于 |
| 1669 | `bool string_starts_with(String const &s, String const &prefix)` | 判断开头 |
| 1675 | `bool string_ends_with(String const &s, String const &suffix)` | 判断结尾 |
| 1681 | `bool string_starts_with(String const &s, u8 prefix)` | 判断单字节开头 |
| 1687 | `bool string_ends_with(String const &s, u8 suffix)` | 判断单字节结尾 |
| 1693 | `String string_trim_starts_with(String const &s, String const &prefix)` | 去除前缀 |
| 1699 | `String string_split_iterator(String_Iterator *it, const char sep)` | 分割迭代器 |
| 1716 | `bool is_separator(u8 const &ch)` | 是否路径分隔符 |
| 1719 | `isize string_extension_position(String const &str)` | 文件扩展名位置 |
| 1732 | `String path_extension(String const &str, bool include_dot)` | 路径扩展名 |
| 1739 | `String path_remove_extension(String const &str)` | 去除扩展名 |
| 1746 | `String string_trim_whitespace(String str)` | 去除空白 |
| 1759 | `String string_trim_trailing_whitespace(String str)` | 去除尾部空白 |
| 1770 | `String split_lines_first_line_from_array(Array<u8> const &array, gbAllocator allocator)` | 首行 |
| 1776 | `Array<String> split_lines_from_array(Array<u8> const &array, gbAllocator allocator)` | 分割行 |

### Rabin-Karp哈希

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1791 | `u32 hash_str_rabin_karp(String const &s, u32 *pow_)` | Rabin-Karp哈希 |

### 字符串查找

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1807 | `isize string_index(String const &s, String const &substr)` | 子串查找 |

### StringPartition

| 行号 | 类型 | 说明 |
|------|------|------|
| 1841-1844 | `StringPartition` | 字符串分区(head, match, tail) |

### StringPartition函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1846 | `StringPartition string_partition(String const &str, String const &sep)` | 分区 |

### 字符串包含

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1858 | `bool string_contains_char(String const &s, u8 c)` | 包含字符 |
| 1866 | `bool string_contains_string(String const &haystack, String const &needle)` | 包含子串 |

### 文件路径函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1883 | `String filename_from_path(String s)` | 文件名(不含扩展名) |
| 1890 | `String filename_without_directory(String s)` | 获取文件名部分 |

### 字符串拼接

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1899 | `String clone_string(gbAllocator a, String const &x)` | 克隆字符串 |
| 1905 | `String concatenate_strings(gbAllocator a, String const &x, String const &y)` | 拼接2个 |
| 1913 | `String concatenate3_strings(gbAllocator a, String const &x, String const &y, String const &z)` | 拼接3个 |
| 1922 | `String concatenate4_strings(gbAllocator a, String const &x, String const &y, String const &z, String const &w)` | 拼接4个 |
| 1932 | `String escape_char(gbAllocator a, String s, char cte)` | 转义字符 |
| 1953 | `String string_join_and_quote(gbAllocator a, Array<String> strings)` | 连接并引用 |
| 1970 | `String copy_string(gbAllocator a, String const &s)` | 复制字符串 |
| 1976 | `String normalize_path(gbAllocator a, String const &path, String const &sep)` | 规范化路径 |

### 字符转换

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 1996 | `int convert_multibyte_to_widechar(...)` | 多字节转宽字符 |
| 1999 | `int convert_widechar_to_multibyte(...)` | 宽字符转多字节 |
| 2002 | `String16 string_to_string16(gbAllocator a, String s)` | String转String16 |
| 2021 | `String string16_to_string(gbAllocator a, String16 s)` | String16转String |

### 临时目录

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2041 | `String temporary_directory(gbAllocator allocator)` | 获取临时目录 |

### 字符判断

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2059 | `bool is_printable(Rune r)` | 是否可打印 |

### 编码相关

| 行号 | 变量 | 说明 |
|------|------|------|
| 2071 | `char const lower_hex[]` | 十六进制字符表 |

### 字符串编码

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2072 | `String quote_to_ascii(gbAllocator a, String str, u8 quote)` | 引用为ASCII |
| 2149 | `String quote_to_ascii(gbAllocator a, String16 str, u8 quote)` | String16版本 |
| 2139 | `Rune decode_surrogate_pair(u16 r1, u16 r2)` | 解码代理对 |

### 字符解码

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2229 | `bool unquote_char(String s, u8 quote, Rune *rune, bool *multiple_bytes, String *tail_string)` | 解码字符 |
| 2339 | `i32 unquote_string(gbAllocator a, String *s_, u8 quote, bool has_carriage_return)` | 解码字符串 |

### 辅助函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2327 | `String strip_carriage_return(gbAllocator a, String s)` | 去除回车符 |
| 2414 | `bool string_is_valid_identifier(String str)` | 是否有效标识符 |

---

## 7. src/range_cache.cpp (行 2439-2488)

### RangeValue 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 2440-2442 | `RangeValue` | 范围值(lo, hi) |
| 2444-2446 | `RangeCache` | 范围缓存 |

### RangeCache 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2447 | `RangeCache range_cache_make(gbAllocator a)` | 创建范围缓存 |
| 2452 | `void range_cache_destroy(RangeCache *c)` | 销毁范围缓存 |
| 2455 | `bool range_cache_add_index(RangeCache *c, i64 index)` | 添加索引 |
| 2466 | `bool range_cache_add_range(RangeCache *c, i64 lo, i64 hi)` | 添加范围 |

---

## 8. 工具函数 (行 2489-2717)

### 幂运算函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2491 | `bool is_power_of_two(i64 x)` | 是否2的幂 |
| 2497 | `bool is_power_of_two_u64(u64 x)` | 是否2的幂(u64) |
| 2511 | `bool is_power_of_two_u64(u64 x)` | 版本1重复 |
| 2553 | `u64 fnv64a(void const *data, isize len, u64 seed)` | FNV哈希 |
| 2535 | `u32 fnv32a(void const *data, isize len)` | FNV32哈希 |

### 比较函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2503 | `int isize_cmp(isize x, isize y)` | isize比较 |
| 2511 | `int u64_cmp(u64 x, u64 y)` | u64比较 |
| 2519 | `int i64_cmp(i64 x, i64 y)` | i64比较 |
| 2527 | `int i32_cmp(i32 x, i32 y)` | i32比较 |

### 数字转换

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2571 | `u64 u64_digit_value(Rune r)` | 数字字符值 |
| 2598 | `u64 u64_from_string(String string)` | 字符串转数字 |

### 数字转字符串

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2637 | `String u64_to_string(u64 v, char *out_buf, isize out_buf_len)` | 数字转字符串 |
| 2650 | `String i64_to_string(i64 a, char *out_buf, isize out_buf_len)` | 有符号数字转字符串 |

### 整数边界常量

| 行号 | 常量 | 说明 |
|------|------|------|
| 2672 | `signed_integer_mins[]` | 有符号整数最小值数组 |
| 2683 | `signed_integer_maxs[]` | 有符号整数最大值数组 |
| 2694 | `unsigned_integer_maxs[]` | 无符号整数最大值数组 |

### 溢出检测

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2705 | `bool add_overflow_u64(u64 x, u64 y, u64 *result)` | 加法溢出检测 |
| 2709 | `bool sub_overflow_u64(u64 x, u64 y, u64 *result)` | 减法溢出检测 |
| 2713 | `void mul_overflow_u64(u64 x, u64 y, u64 *lo, u64 *hi)` | 乘法溢出检测 |

### 全局变量

| 行号 | 变量 | 说明 |
|------|------|------|
| 2716 | `global_module_path` | 全局模块路径 |
| 2717 | `global_module_path_set` | 模块路径是否已设置 |

---

## 9. src/ptr_map.cpp (行 2718-3071)

### PtrMap 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 2719 | `MapIndex` | 地图索引(u32) |
| 2720-2723 | 缓存行大小常量 | MAP_CACHE_LINE_SIZE相关 |
| 2725 | `MAP_SENTINEL` | 哨兵索引 |
| 2727-2737 | `PtrMapConstant<T>` | 指针映射常量(指针版本) |
| 2739-2743 | `PtrMapConstant<T>` | 指针映射常量(i64版本) |
| 2746-2750 | `PtrMapEntry<K, V>` | 指针映射条目 |
| 2752-2756 | `PtrMap<K, V>` | 指针映射 |

### PtrMap 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 2757 | `u32 ptr_map_hash_key(uintptr key)` | 哈希键函数 |
| 2768 | `u32 ptr_map_hash_key(void const *key)` | 指针版本 |
| 2771 | `void map_init(PtrMap<K, V> *h, isize capacity)` | 初始化 |
| 2796 | `void map_destroy(PtrMap<K, V> *h)` | 销毁 |
| 2802 | `void map__insert(PtrMap<K, V> *h, K key, V const &value)` | 内部插入 |
| 2823 | `b32 map__full(PtrMap<K, V> *h)` | 是否满 |
| 2827 | `void map_grow(PtrMap<K, V> *h)` | 扩容 |
| 2837 | `void try_map_grow(PtrMap<K, V> *h)` | 尝试扩容 |
| 2843 | `void map_reserve(PtrMap<K, V> *h, isize cap)` | 预留 |
| 2861 | `V *map_get(PtrMap<K, V> *h, K key)` | 获取 |
| 2884 | `V *map_try_get(PtrMap<K, V> *h, K key, MapIndex *found_index_)` | 尝试获取 |
| 2909 | `void map_set_internal_from_try_get(PtrMap<K, V> *h, K key, V const &value, MapIndex found_index)` | 从尝试获取设置 |
| 2918 | `V &map_must_get(PtrMap<K, V> *h, K key)` | 必须获取 |
| 2924 | `void map_set(PtrMap<K, V> *h, K key, V const &value)` | 设置 |
| 2935 | `bool map_set_if_not_previously_exists(PtrMap<K, V> *h, K key, V const &value)` | 不存在则设置 |
| 2945 | `void map_remove(PtrMap<K, V> *h, K key)` | 移除 |
| 2953 | `void map_clear(PtrMap<K, V> *h)` | 清空 |
| 2958 | `PtrMapEntry<K, V> *multi_map_find_first(PtrMap<K, V> *h, K key)` | 多映射查找首个 |
| 2978 | `PtrMapEntry<K, V> *multi_map_find_next(PtrMap<K, V> *h, PtrMapEntry<K, V> *e)` | 多映射查找下一个 |
| 2995 | `isize multi_map_count(PtrMap<K, V> *h, K key)` | 多映射计数 |
| 3005 | `void multi_map_get_all(PtrMap<K, V> *h, K key, V *items)` | 多映射获取全部 |
| 3014 | `void multi_map_insert(PtrMap<K, V> *h, K key, V const &value)` | 多映射插入 |
| 3019 | `void multi_map_remove_all(PtrMap<K, V> *h, K key)` | 多映射移除全部 |

### OrderedInsertPtrMap 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 3087-3089 | `MapFindResult` | 地图查找结果 |
| 3092-3098 | `OrderedInsertPtrMapEntry<K, V>` | 有序插入映射条目 |
| 3100-3106 | `OrderedInsertPtrMap<K, V>` | 有序插入映射 |

### OrderedInsertPtrMap 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 3124 | `void map_init(OrderedInsertPtrMap<K, V> *h, isize capacity)` | 初始化(有序) |
| 3129 | `void map_destroy(OrderedInsertPtrMap<K, V> *h)` | 销毁(有序) |
| 3135 | `void map__resize_hashes(OrderedInsertPtrMap<K, V> *h, usize count)` | 调整哈希大小 |
| 3139 | `void map__reserve_entries(OrderedInsertPtrMap<K, V> *h, usize capacity)` | 预留条目 |
| 3143 | `MapIndex map__add_entry(OrderedInsertPtrMap<K, V> *h, K key)` | 添加条目 |
| 3154 | `MapFindResult map__find(OrderedInsertPtrMap<K, V> *h, K key)` | 查找 |
| 3173 | `MapFindResult map__find_from_entry(OrderedInsertPtrMap<K, V> *h, OrderedInsertPtrMapEntry<K, V> *e)` | 从条目查找 |
| 3191 | `b32 map__full(OrderedInsertPtrMap<K, V> *h)` | 是否满 |
| 3195 | `void map_grow(OrderedInsertPtrMap<K, V> *h)` | 扩容 |
| 3200 | `void map_reset_entries(OrderedInsertPtrMap<K, V> *h)` | 重置条目 |
| 3217 | `void map_reserve(OrderedInsertPtrMap<K, V> *h, isize cap)` | 预留 |
| 3226 | `void map_rehash(OrderedInsertPtrMap<K, V> *h, isize new_count)` | 重哈希 |
| 3230 | `V *map_get(OrderedInsertPtrMap<K, V> *h, K key)` | 获取 |
| 3250 | `V *map_try_get(OrderedInsertPtrMap<K, V> *h, K key, MapFindResult *fr_)` | 尝试获取 |
| 3272 | `void map_set_internal_from_try_get(OrderedInsertPtrMap<K, V> *h, K key, V const &value, MapFindResult const &fr)` | 设置 |
| 3282 | `V &map_must_get(OrderedInsertPtrMap<K, V> *h, K key)` | 必须获取 |
| 3288 | `void map_set(OrderedInsertPtrMap<K, V> *h, K key, V const &value)` | 设置 |
| 3311 | `bool map_set_if_not_previously_exists(OrderedInsertPtrMap<K, V> *h, K key, V const &value)` | 不存在则设置 |
| 3335 | `void map__erase(OrderedInsertPtrMap<K, V> *h, MapFindResult const &fr)` | 删除 |
| 3356 | `void map_remove(OrderedInsertPtrMap<K, V> *h, K key)` | 移除 |
| 3363 | `void map_clear(OrderedInsertPtrMap<K, V> *h)` | 清空 |
| 3370 | `OrderedInsertPtrMapEntry<K, V> *multi_map_find_first(OrderedInsertPtrMap<K, V> *h, K key)` | 多映射查找首个 |
| 3378 | `OrderedInsertPtrMapEntry<K, V> *multi_map_find_next(OrderedInsertPtrMap<K, V> *h, OrderedInsertPtrMapEntry<K, V> *e)` | 多映射查找下一个 |
| 3389 | `isize multi_map_count(OrderedInsertPtrMap<K, V> *h, K key)` | 多映射计数 |
| 3399 | `void multi_map_get_all(OrderedInsertPtrMap<K, V> *h, K key, V *items)` | 多映射获取全部 |
| 3408 | `void multi_map_insert(OrderedInsertPtrMap<K, V> *h, K key, V const &value)` | 多映射插入 |
| 3428 | `void multi_map_remove(OrderedInsertPtrMap<K, V> *h, K key, OrderedInsertPtrMapEntry<K, V> *e)` | 多映射移除 |
| 3435 | `void multi_map_remove_all(OrderedInsertPtrMap<K, V> *h, K key)` | 多映射移除全部 |

### 迭代器

| 行号 | 类型 | 说明 |
|------|------|------|
| 3025-3047 | `PtrMapIterator<K, V>` | 指针映射迭代器 |

### 迭代器函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 3049 | `PtrMapIterator<K, V> end(PtrMap<K, V> &m)` | 结束迭代器 |
| 3053 | `PtrMapIterator<K, V> const end(PtrMap<K, V> const &m)` | 结束迭代器(const) |
| 3057 | `PtrMapIterator<K, V> begin(PtrMap<K, V> &m)` | 起始迭代器 |
| 3057 | `PtrMapIterator<K, V> const begin(PtrMap<K, V> const &m)` | 起始迭代器(const) |

---

## 10. src/ptr_set.cpp (行 3456-3628)

### PtrSet 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 3457 | `PTR_SET_INLINE_CAP` | 内联容量(16) |
| 3459-3466 | `PtrSet<T>` | 指针集合 |

### PtrSet 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 3474 | `gbAllocator ptr_set_allocator(void)` | 获取分配器 |
| 3478 | `void ptr_set_init(PtrSet<T> *s, isize capacity)` | 初始化 |
| 3492 | `void ptr_set_destroy(PtrSet<T> *s)` | 销毁 |
| 3501 | `isize ptr_set__find(PtrSet<T> *s, T ptr)` | 查找 |
| 3520 | `bool ptr_set__full(PtrSet<T> *s)` | 是否满 |
| 3525 | `void ptr_set_grow(PtrSet<T> *old_set)` | 扩容 |
| 3541 | `bool ptr_set_exists(PtrSet<T> *s, T ptr)` | 存在检查 |
| 3545 | `bool ptr_set_update(PtrSet<T> *s, T ptr)` | 更新 |
| 3575 | `bool ptr_set_update_with_mutex(PtrSet<T> *s, T ptr, RWSpinLock *m)` | 带互斥更新 |
| 3609 | `T ptr_set_add(PtrSet<T> *s, T ptr)` | 添加 |
| 3614 | `void ptr_set_remove(PtrSet<T> *s, T ptr)` | 移除 |
| 3625 | `void ptr_set_clear(PtrSet<T> *s)` | 清空 |

---

## 11. src/string_map.cpp (行 3629-3888)

### StringMap 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 3631 | `StringHashKey` | 字符串哈希键 |
| 3641 | `u32 string_hash(String const &s)` | 字符串哈希 |
| 3645 | `StringHashKey string_hash_string(String const &s)` | 生成哈希键 |
| 3652-3657 | `StringMapEntry<T>` | 字符串映射条目 |
| 3659-3665 | `StringMap<T>` | 字符串映射 |

### StringMap 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 3680 | `gbAllocator string_map_allocator(void)` | 获取分配器 |
| 3684 | `void string_map_init(StringMap<T> *h, usize capacity)` | 初始化 |
| 3689 | `void string_map_destroy(StringMap<T> *h)` | 销毁 |
| 3694 | `void string_map__resize_hashes(StringMap<T> *h, usize count)` | 调整大小 |
| 3698 | `void string_map__reserve_entries(StringMap<T> *h, usize capacity)` | 预留条目 |
| 3702 | `MapIndex string_map__add_entry(StringMap<T> *h, u32 hash, String const &key)` | 添加条目 |
| 3714 | `MapFindResult string_map__find(StringMap<T> *h, u32 hash, String const &key)` | 查找 |
| 3731 | `MapFindResult string_map__find_from_entry(StringMap<T> *h, StringMapEntry<T> *e)` | 从条目查找 |
| 3748 | `b32 string_map__full(StringMap<T> *h)` | 是否满 |
| 3752 | `void string_map_grow(StringMap<T> *h)` | 扩容 |
| 3757 | `void string_map_reset_entries(StringMap<T> *h)` | 重置条目 |
| 3774 | `void string_map_reserve(StringMap<T> *h, usize cap)` | 预留 |
| 3783 | `T *string_map_get(StringMap<T> *h, u32 hash, String const &key)` | 获取 |
| 3800 | `T *string_map_get(StringMap<T> *h, StringHashKey const &key)` | 获取(哈希键) |
| 3804 | `T *string_map_get(StringMap<T> *h, String const &key)` | 获取(字符串) |
| 3808 | `T *string_map_get(StringMap<T> *h, char const *key)` | 获取(C字符串) |
| 3813 | `T &string_map_must_get(StringMap<T> *h, u32 hash, String const &key)` | 必须获取 |
| 3819 | `T &string_map_must_get(StringMap<T> *h, StringHashKey const &key)` | 必须获取(哈希键) |
| 3823 | `T &string_map_must_get(StringMap<T> *h, String const &key)` | 必须获取(字符串) |
| 3827 | `T &string_map_must_get(StringMap<T> *h, char const *key)` | 必须获取(C字符串) |
| 3832 | `void string_map_set(StringMap<T> *h, u32 hash, String const &key, T const &value)` | 设置 |
| 3855 | `void string_map_set(StringMap<T> *h, String const &key, T const &value)` | 设置(字符串) |
| 3859 | `void string_map_set(StringMap<T> *h, char const *key, T const &value)` | 设置(C字符串) |
| 3863 | `void string_map_set(StringMap<T> *h, StringHashKey const &key, T const &value)` | 设置(哈希键) |
| 3867 | `void string_map_clear(StringMap<T> *h)` | 清空 |

### 迭代器

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 3874 | `StringMapEntry<T> *begin(StringMap<T> &m)` | 起始��代�� |
| 3878 | `StringMapEntry<T> const *begin(StringMap<T> const &m)` | 起始迭代器(const) |
| 3882 | `StringMapEntry<T> *end(StringMap<T> &m)` | 结束迭代器 |
| 3886 | `StringMapEntry<T> const *end(StringMap<T> const &m)` | 结束迭代器(const) |

---

## 12. src/string16_map.cpp (行 3890-4118)

### String16Map 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 3892 | `String16HashKey` | 字符串16哈希键 |
| 3902 | `u32 string_hash(String16 const &s)` | 字符串16哈希 |
| 3906 | `String16HashKey string_hash_string(String16 const &s)` | 生成哈希键 |
| 3913-3917 | `String16MapEntry<T>` | 字符串16映射条目 |
| 3919-3926 | `String16Map<T>` | 字符串16映射 |

### String16Map 函数

类似StringMap函数，只是操作String16类型。

---

## 13. src/string_set.cpp (行 4119-4339)

### StringSet 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 4120-4124 | `StringSetEntry` | 字符串集合条目 |
| 4131-4134 | `StringSet` | 字符串集合 |

### StringSet 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4144 | `gbAllocator string_set_allocator(void)` | 获取分配器 |
| 4147 | `void string_set_init(StringSet *s, isize capacity)` | 初始化 |
| 4155 | `void string_set_destroy(StringSet *s)` | 销毁 |
| 4162 | `MapIndex string_set__add_entry(StringSet *s, StringHashKey const &key)` | 添加条目 |
| 4170 | `MapFindResult string_set__find(StringSet *s, StringHashKey const &key)` | 查找 |
| 4186 | `MapFindResult string_set__find_from_entry(StringSet *s, StringSetEntry *e)` | 从条目查找 |
| 4201 | `b32 string_set__full(StringSet *s)` | 是否满 |
| 4204 | `void string_set_grow(StringSet *s)` | 扩容 |
| 4208 | `void string_set_reset_entries(StringSet *s)` | 重置条目 |
| 4224 | `void string_set_reserve(StringSet *s, isize cap)` | 预留 |
| 4235 | `void string_set_rehash(StringSet *s, isize new_count)` | 重哈希 |
| 4238 | `bool string_set_exists(StringSet *s, String const &str)` | 存在检查 |
| 4243 | `void string_set_add(StringSet *s, String const &str)` | 添加 |
| 4266 | `bool string_set_update(StringSet *s, String const &str)` | 更新 |
| 4292 | `void string_set__erase(StringSet *s, MapFindResult fr)` | 删除 |
| 4315 | `void string_set_remove(StringSet *s, String const &str)` | 移除 |
| 4322 | `void string_set_clear(StringSet *s)` | 清空 |

### 迭代器

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4328 | `StringSetEntry *begin(StringSet &m)` | 起始迭代器 |
| 4334 | `StringSetEntry *end(StringSet &m)` | 结束迭代器 |

---

## 14. src/priority_queue.cpp (行 4340-4420)

### PriorityQueue 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 4341-4346 | `PriorityQueue<T>` | 优先队列 |

### PriorityQueue 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4348 | `bool priority_queue_shift_down(PriorityQueue<T> *pq, isize i0, isize n)` | 下沉 |
| 4367 | `void priority_queue_shift_up(PriorityQueue<T> *pq, isize j)` | 上浮 |
| 4378 | `void priority_queue_fix(PriorityQueue<T> *pq, isize i)` | 修复堆 |
| 4384 | `void priority_queue_push(PriorityQueue<T> *pq, T const &value)` | 入队 |
| 4389 | `T priority_queue_pop(PriorityQueue<T> *pq)` | 出队 |
| 4397 | `T priority_queue_remove(PriorityQueue<T> *pq, isize i)` | 移除指定位置 |
| 4408 | `PriorityQueue<T> priority_queue_create(...)` | 创建优先队列 |

---

## 15. src/thread_pool.cpp (行 4421-4604)

### 线程池类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 4422 | `WorkerTask` | 工作任务 |
| 4423 | `ThreadPool` | 线程池 |

### GrabState 枚举

| 行号 | 枚举值 | 说明 |
|------|--------|------|
| 4432 | `Grab_Success` | 0 - 成功 |
| 4433 | `Grab_Empty` | 1 - 空 |
| 4434 | `Grab_Failed` | 2 - 失败 |

### BroadcastWaitState 枚举

| 行号 | 枚举值 | 说明 |
|------|--------|------|
| 4437 | `Nobody_Waiting` | 0 - 无人等待 |
| 4438 | `Someone_Waiting` | 1 - 有人等待 |

### 线程池函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4425 | `Thread *get_current_thread(void)` | 获取当前线程 |
| 4428 | `void thread_pool_init(ThreadPool *pool, isize worker_count, char const *worker_name)` | 初始化线程池 |
| 4429 | `void thread_pool_destroy(ThreadPool *pool)` | 销毁线程池 |
| 4430 | `bool thread_pool_add_task(ThreadPool *pool, WorkerTaskProc *proc, void *data)` | 添加任务 |
| 4431 | `void thread_pool_wait(ThreadPool *pool)` | 等待完成 |
| 4448 | `isize current_thread_index(void)` | 当前线程索引 |
| 4472 | `TaskRingBuffer *task_ring_grow(TaskRingBuffer *ring, isize bottom, isize top)` | 环形缓冲扩容 |
| 4479 | `void thread_pool_queue_push(Thread *thread, WorkerTask task)` | 队列推入 |
| 4498 | `GrabState thread_pool_queue_take(Thread *thread, WorkerTask *task)` | 队列取出 |
| 4521 | `GrabState thread_pool_queue_steal(Thread *thread, WorkerTask *task)` | 队列偷取 |
| 4559 | `isize thread_pool_thread_proc(struct Thread *thread)` | 线程处理函数 |

---

## 16. src/string_interner.cpp (行 4605-4810)

### InternedString 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 4606-4614 | `InternedString` | 驻留字符串 |
| 4616-4620 | `StringInternCell` | 驻留字符串单元格(8个哈希槽) |
| 4621-4623 | `PaddedMutex` | 对齐互斥锁 |
| 4624-4626 | `PaddedI64` | 对齐i64 |
| 4627-4635 | `StringInterner` | 字符串驻留器 |

### StringInterner 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4636 | `StringInterner *string_interner_create()` | 创建驻留器 |
| 4637 | `InternedString string_interner_insert(String str, u32 hash, u32 *new_hash_)` | 插入驻留字符串 |
| 4638 | `String string_interner_load(InternedString interned)` | 加载驻留字符串 |

### 全局变量

| 行号 | 变量 | 说明 |
|------|------|------|
| 4639 | `g_string_interner` | 全局字符串驻留器 |
| 4640 | `g_interned_blank_ident` | 空白标识符 |

### StringInternerThreadLocalArena

| 行号 | 类型 | 说明 |
|------|------|------|
| 4641-4644 | `StringInternerThreadLocalArena` | 线程本地Arena |
| 4645 | `g_interner_arena` | 线程本地全局Arena |

### 驻留器函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4648 | `void init_string_interner()` | 初始化驻留器 |
| 4661 | `String string_interner_load(InternedString interned)` | 加载(重载) |
| 4672 | `char const *string_interner_load_cstring(InternedString interned)` | 加载C字符串 |
| 4694 | `InternedString string_interner_insert(...)` | 插入(重载) |
| 4763 | `char const *string_intern_cstring(String str, u32 *hash_)` | 驻留C字符串 |
| 4767 | `String string_intern_string(String str, u32 *hash_)` | 驻留字符串 |
| 4771 | `void string_interner_thread_local_arena_init(StringInternerThreadLocalArena *tl_arena)` | 初始化线程本地Arena |
| 4777 | `void *string_interner_thread_local_arena_alloc(...)` | 分配线程本地Arena |

### 混淆函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4794 | `String obfuscate_string(String const &s, char const *prefix)` | 字符串混淆 |
| 4804 | `i32 obfuscate_i32(i32 i)` | i32混淆 |

---

## 17. 常用工具函数 (行 4811-4940)

### next_pow2 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4811 | `i32 next_pow2(i32 n)` | i32版本 |
| 4824 | `i64 next_pow2(i64 n)` | i64版本 |
| 4838 | `isize next_pow2_isize(isize n)` | isize版本 |
| 4852 | `u32 next_pow2_u32(u32 n)` | u32版本 |

### 位操作函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4865 | `i32 bit_set_count(u32 x)` | 设置位数(i32) |
| 4873 | `i64 bit_set_count(u64 x)` | 设置位数(i64) |
| 4878 | `u32 floor_log2(u32 x)` | 向下对数2 |
| 4886 | `u64 floor_log2(u64 x)` | 向下对数2(u64) |
| 4895 | `u32 ceil_log2(u32 x)` | 向上对数2 |
| 4906 | `u64 ceil_log2(u64 x)` | 向上对数2(u64) |
| 4918 | `u32 prev_pow2(u32 n)` | 前一个2的幂 |
| 4929 | `i32 prev_pow2(i32 n)` | 前一个2的幂(i32) |
| 4940 | `i64 prev_pow2(i64 n)` | 前一个2的幂(i64) |

---

## 18. 浮点数转换 (行 4952-5007)

### 浮点转换函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 4952 | `u16 f32_to_f16(f32 value)` | f32转f16 |
| 4992 | `f32 f16_to_f32(u16 value)` | f16转f32 |
| 5005 | `f64 gb_sqrt(f64 x)` | 平方根 |

---

## 19. src/path.cpp (行 5063-5528)

### 文件路径函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5064 | `String remove_extension_from_path(String const &s)` | 移除扩展名 |
| 5075 | `String remove_directory_from_path(String const &s)` | 移除目录 |
| 5087 | `String get_working_directory(gbAllocator allocator)` | 获取工作目录 |
| 5104 | `bool set_working_directory(String dir)` | 设置工作目录 |
| 5115 | `String directory_from_path(String const &s)` | 获取目录部分 |
| 5131 | `bool path_is_directory(String path)` | 是否目录 |
| 5139 | `String path_to_full_path(gbAllocator a, String path)` | 转完整路径 |

### Path 结构

| 行号 | 类型 | 说明 |
|------|------|------|
| 5152-5156 | `Path` | 路径结构(basename, name, ext) |

### Path 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5157 | `String path_to_string(gbAllocator a, Path path)` | Path转字符串 |
| 5179 | `String quote_path(gbAllocator a, Path path)` | 引用路径 |
| 5185 | `String path_to_full_path(gbAllocator a, Path path)` | Path转完整路径 |
| 5190 | `Path path_from_string(gbAllocator a, String const &path)` | 字符串转Path |
| 5211 | `String last_path_element(String const &path)` | 最后路径元素 |
| 5225 | `bool path_is_directory(Path path)` | 是否目录(Path) |

### FileInfo 结构

| 行号 | 类型 | 说明 |
|------|------|------|
| 5230-5235 | `FileInfo` | 文件信息(name, fullpath, size, is_dir) |

### ReadDirectoryError 枚举

| 行号 | 枚举值 | 说明 |
|------|--------|------|
| 5236-5245 | 错误代码 | 无效路径, 不存在, 权限, 非目录, 空等 |

### 文件操作函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5246 | `i64 get_file_size(String path)` | 获取文件大小 |
| 5257 | `ReadDirectoryError read_directory(String path, Array<FileInfo> *fi)` | 读取目录 |
| 5329 | `bool write_directory(String path)` | ��建��录 |
| 5363 | `LoadedFileError load_file_32(...)` | 加载文件(32位) |

### LoadedFile 结构

| 行号 | 类型 | 说明 |
|------|------|------|
| 5349-5353 | `LoadedFile` | 已加载文件(handle, data, size) |

### LoadedFileError 枚举

| 行号 | 枚举值 | 说明 |
|------|--------|------|
| 5354-5362 | 错误代码 | 无, 空, 文件过大, 无效, 不存在, 权限等 |

### 距离函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5448 | `isize levenstein_distance_case_insensitive(String const &a, String const &b)` | Levenshtein距离(不区分大小写) |

### DidYouMean 类型

| 行号 | 类型 | 说明 |
|------|------|------|
| 5488-5490 | `DistanceAndTarget` | 距离和目标 |
| 5492-5495 | `DidYouMeanAnswers` | 建议答案 |

### DidYouMean 常量

| 行号 | 常量 | 值 |
|------|------|-----|
| 5496 | `MAX_SMALLEST_DID_YOU_MEAN_DISTANCE` | 3-1 = 2 |

### DidYouMean 函数

| 行号 | 函数签名 | 说明 |
|------|----------|------|
| 5497 | `DidYouMeanAnswers did_you_mean_make(gbAllocator allocator, isize cap, String const &key)` | 创建建议器 |
| 5503 | `void did_you_mean_destroy(DidYouMeanAnswers *d)` | 销毁 |
| 5506 | `void did_you_mean_append(DidYouMeanAnswers *d, String const &target)` | 添加目标 |
| 5515 | `Slice<DistanceAndTarget> did_you_mean_results(DidYouMeanAnswers *d)` | 获取结果 |

---

## 总结

本文档涵盖了 `common.i.cpp` 中的所有主要类型和函数，包括：

1. **数组/切片操作** - Array, Slice 的创建、添加、删除、扩容等
2. **线程与同步** - Mutex, RWMutex, Semaphore, Parker, Thread, ThreadPool
3. **内存管理** - Arena, StaticArena, MemoryBlock
4. **队列** - MPSCQueue, MPMCQueue
5. **字符串操作** - String, String16 的创建、比较、搜索、分割
6. **映射/集合** - PtrMap, StringMap, PtrSet, StringSet
7. **优先队列** - PriorityQueue
8. **路径操作** - 文件路径解析、目录操作
9. **工具函数** - 位操作、哈希、溢出检测等

每个函数都标注了对应的源码行号，方便追溯和理解。