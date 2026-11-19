# 内置词库

内置词库在编译期通过 `go:embed` 嵌入，可直接传入 `LoadDictEmbed` 使用。词库以两个上游工程的公开词库为基础合并、去空行、按小写去重，并并入 `LuYongwang/go-sensitive-word` 同名分类中的额外词条。

| 常量 | 文件 | 词数 |
|------|------|------|
| `DictCovid19` | COVID-19词库.txt | 72 |
| `DictGFWAdditional` | GFW补充词库.txt | 6169 |
| `DictOther` | 其他词库.txt | 153 |
| `DictReactionary` | 反动词库.txt | 551 |
| `DictAdvertisement` | 广告类型.txt | 120 |
| `DictPolitical` | 政治类型.txt | 303 |
| `DictViolence` | 暴恐词库.txt | 178 |
| `DictPeopleLife` | 民生词库.txt | 510 |
| `DictGunExplosion` | 涉枪涉爆.txt | 434 |
| `DictNeteaseFE` | 网易前端过滤敏感词库.txt | 7746 |
| `DictSexual` | 色情类型.txt | 304 |
| `DictPornography` | 色情词库.txt | 587 |
| `DictAdditional` | 补充词库.txt | 1063 |
| `DictCorruption` | 贪腐词库.txt | 240 |
| `DictTemporaryTencent` | 零时-Tencent.txt | 41588 |
| `DictIllegalURL` | 非法网址.txt | 14592 |

全量 16 个词库去重并集约 **51,103** 词。其中 `零时-Tencent`、`非法网址`、`网易前端过滤敏感词库`、`GFW补充词库` 体积较大，建议按需加载。

## 来源

- `zmexing/go-sensitive-word`：16 个分类词库的基础。
- `LuYongwang/go-sensitive-word`：同名分类中额外的 35 个色情词，以及去重整理后的分类集合。
- 原始词库参考 `konsheng/Sensitive-lexicon`。

## 使用

```go
err := filter.LoadDictEmbed(
    sensitive.DictPolitical,
    sensitive.DictViolence,
    sensitive.DictPornography,
)
```
