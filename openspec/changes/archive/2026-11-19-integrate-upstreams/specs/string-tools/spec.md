# Spec Delta

## Purpose

提供邮箱地址、URL 网址、数字字符与微信号等常见敏感信息的检测与脱敏工具函数，使调用方无需自行编写正则表达式即可完成基础内容治理。

## ADDED Requirements

### Requirement: 邮箱检测与脱敏
系统 SHALL 判断字符串中是否存在邮箱地址，并 SHALL 将匹配到的邮箱地址整体替换为固定掩码。

#### Scenario: 检测邮箱
- **WHEN** 输入包含 `user@example.com`
- **THEN** `HasEmail` 返回真

#### Scenario: 脱敏邮箱
- **WHEN** 输入包含 `user@example.com`
- **THEN** `MaskEmail` 返回的字符串不再包含原邮箱，原位置被固定掩码替代

#### Scenario: 无邮箱
- **WHEN** 输入不包含邮箱地址
- **THEN** `HasEmail` 返回假，`MaskEmail` 原样返回输入

### Requirement: URL 检测与脱敏
系统 SHALL 判断字符串中是否存在 HTTP/HTTPS 网址，并 SHALL 将匹配到的网址整体替换为固定掩码。

#### Scenario: 检测网址
- **WHEN** 输入包含 `https://example.com/path`
- **THEN** `HasURL` 返回真

#### Scenario: 脱敏网址
- **WHEN** 输入包含网址
- **THEN** `MaskURL` 返回的字符串不再包含原网址，原位置被固定掩码替代

### Requirement: 数字检测与脱敏
系统 SHALL 统计字符串中的数字字符数量，并在达到阈值时返回真；SHALL 将每个数字字符替换为掩码字符。

#### Scenario: 达到数字阈值
- **WHEN** 输入包含至少 `count` 个数字字符
- **THEN** `HasDigit(s, count)` 返回真

#### Scenario: 未达到数字阈值
- **WHEN** 输入中的数字字符少于 `count`
- **THEN** `HasDigit(s, count)` 返回假

#### Scenario: 脱敏数字
- **WHEN** 输入包含数字
- **THEN** `MaskDigit` 将每个数字字符替换为 `*`，非数字字符保持不变

### Requirement: 微信号检测与脱敏
系统 SHALL 判断字符串中是否存在微信号格式的子串，并 SHALL 将匹配到的微信号整体替换为固定掩码。

#### Scenario: 检测微信号
- **WHEN** 输入包含符合微信号格式的子串
- **THEN** `HasWechatID` 返回真

#### Scenario: 脱敏微信号
- **WHEN** 输入包含符合微信号格式的子串
- **THEN** `MaskWechatID` 返回的字符串不再包含原微信号，原位置被固定掩码替代
