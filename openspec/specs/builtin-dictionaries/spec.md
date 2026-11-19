# builtin-dictionaries Specification

## Purpose
在编译期嵌入多分类敏感词库，并以稳定的导出常量统一暴露；调用方无需额外文件、网络或运行时下载，即可按业务场景选择性地加载所需分类词库并立即用于检测。

## Requirements

### Requirement: 内置词库常量
系统 SHALL 提供以下 16 个非空的词库常量，供嵌入式加载使用：`DictCovid19`、`DictGFWAdditional`、`DictOther`、`DictReactionary`、`DictAdvertisement`、`DictPolitical`、`DictViolence`、`DictPeopleLife`、`DictGunExplosion`、`DictNeteaseFE`、`DictSexual`、`DictPornography`、`DictAdditional`、`DictCorruption`、`DictTemporaryTencent`、`DictIllegalURL`。

#### Scenario: 常量均可用
- **WHEN** 调用方引用上述任一词库常量
- **THEN** 编译成功且常量内容非空

#### Scenario: 按需加载
- **WHEN** 调用方仅加载 `DictPolitical`
- **THEN** 其余词库的词条不会自动进入词库

### Requirement: 词库内容规范
每个内置词库 SHALL 为按行分隔的纯文本，每行一个词条，MUST 去除空行与重复词条。

#### Scenario: 无空行与重复
- **WHEN** 逐行读取任一内置词库常量
- **THEN** 不存在空行，且不存在重复词条

#### Scenario: 同名分类取并集
- **WHEN** 两个上游工程存在同名分类词库
- **THEN** 内置版本包含两方该分类词条的去重并集

### Requirement: 与加载接口配合
内置词库常量 SHALL 可直接传入嵌入式加载接口，加载后词条可被检测 API 命中。

#### Scenario: 加载后命中
- **WHEN** 调用方加载 `DictViolence` 并在文本中写入其中任一非空词条
- **THEN** `IsSensitive` 返回真，`FindAll` 包含该词条
