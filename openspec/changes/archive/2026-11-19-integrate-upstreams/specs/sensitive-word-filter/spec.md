# Spec Delta

## Purpose

提供基于 DFA 或 AC 自动机的敏感词检测、查找、替换与删除能力，并保证在词库动态变化和高并发查询下的行为可预期。

## ADDED Requirements

### Requirement: 算法选择
系统 SHALL 允许在创建过滤器时选择 DFA 或 AC 算法，两种算法对同一词库和文本的检测结果 MUST 等价。

#### Scenario: 使用 DFA
- **WHEN** 调用方以 `FilterDfa` 创建过滤器并加载词库
- **THEN** 所有检测、替换与删除 API 均可正常调用

#### Scenario: 使用 AC
- **WHEN** 调用方以 `FilterAc` 或 `FilterAC` 创建过滤器并加载词库
- **THEN** 所有检测、替换与删除 API 均可正常调用，且结果与 DFA 一致

#### Scenario: 非法算法类型
- **WHEN** 传入未定义的过滤器类型
- **THEN** 创建过滤器返回错误，不产生可用的过滤器实例

### Requirement: 敏感词检测
系统 SHALL 提供判断是否存在敏感词、查找首个敏感词、查找全部敏感词（去重）以及统计出现次数的能力。

#### Scenario: 存在敏感词
- **WHEN** 文本包含已加载的词
- **THEN** `IsSensitive` 返回真，`FindOne` 返回命中的首个词，`FindAll` 包含该词

#### Scenario: 不存在敏感词
- **WHEN** 文本不包含任何已加载的词
- **THEN** `IsSensitive` 返回假，`FindOne` 返回空字符串，`FindAll` 返回空集合

#### Scenario: 重复词去重与计数
- **WHEN** 同一敏感词在文本中出现多次
- **THEN** `FindAll` 只返回一次该词，`FindAllCount` 返回其实际出现次数

### Requirement: 替换与删除
系统 SHALL 支持将命中的敏感词替换为指定字符，或从文本中完全移除命中的敏感词。

#### Scenario: 替换敏感词
- **WHEN** 调用 `Replace` 并指定替换字符
- **THEN** 每个命中词覆盖的原文位置均被该字符替换，未命中文本保持不变

#### Scenario: 删除敏感词
- **WHEN** 调用 `Remove`
- **THEN** 所有命中词被移除，其余文本按原顺序保留

### Requirement: 匹配结果返回原文
系统 SHALL 在查找、替换与删除结果中基于原始文本处理，MUST NOT 返回归一化后的替代字符。

#### Scenario: 大小写混合文本
- **WHEN** 词库包含小写词而文本以大写形式出现
- **THEN** 查找结果返回原文中的大写片段，替换与删除作用于原文对应字符

### Requirement: 并发安全
系统 SHALL 支持并发查询，同时在词库持续新增或删除词条时不发生数据竞争或 panic。

#### Scenario: 查询期间更新词库
- **WHEN** 多个 goroutine 并发调用检测 API，另一 goroutine 同时添加或删除词条
- **THEN** 所有调用安全返回，竞态检测器不报告数据竞争

### Requirement: 更新可见性
系统 SHALL 在词库更新后于有限时间内使后续查询可见，并 MUST 提供同步点以确定性等待待处理更新完成。

#### Scenario: 等待同步后查询
- **WHEN** 调用方添加词条后调用同步等待
- **THEN** 同步返回后，后续查询必定能看到该词条

#### Scenario: 有限时间窗口内可见
- **WHEN** 调用方添加词条后不调用同步等待，并在 200 毫秒内重复查询
- **THEN** 该词条最终可被检测到
