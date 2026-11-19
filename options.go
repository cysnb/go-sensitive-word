// Package go_sensitive_word 提供基于 DFA/AC 的敏感词过滤能力。
package go_sensitive_word

import _ "embed"

// StoreMemory 内存存储类型。
const (
	StoreMemory = iota
)

// 过滤算法类型。
const (
	FilterDfa = iota
	FilterAc
)

// FilterAC 是 FilterAc 的兼容别名。
const FilterAC = FilterAc

// StoreOption 词库存储配置。
type StoreOption struct {
	Type uint32
}

// FilterOption 过滤算法配置。
type FilterOption struct {
	Type uint32
}

type config struct {
	normalizer NormalizerConfig
}

// Option 过滤器可选配置。
type Option func(*config)

// WithNormalizer 指定自定义归一化配置。
func WithNormalizer(cfg NormalizerConfig) Option {
	return func(c *config) {
		c.normalizer = cfg
	}
}

// WithStrictNormalizer 使用严格归一化配置。
func WithStrictNormalizer() Option {
	return func(c *config) {
		c.normalizer = StrictNormalizer()
	}
}

// 内置敏感词库（编译期嵌入）。
var (
	//go:embed wordlists/COVID-19词库.txt
	DictCovid19 string
	//go:embed wordlists/GFW补充词库.txt
	DictGFWAdditional string
	//go:embed wordlists/其他词库.txt
	DictOther string
	//go:embed wordlists/反动词库.txt
	DictReactionary string
	//go:embed wordlists/广告类型.txt
	DictAdvertisement string
	//go:embed wordlists/政治类型.txt
	DictPolitical string
	//go:embed wordlists/暴恐词库.txt
	DictViolence string
	//go:embed wordlists/民生词库.txt
	DictPeopleLife string
	//go:embed wordlists/涉枪涉爆.txt
	DictGunExplosion string
	//go:embed wordlists/网易前端过滤敏感词库.txt
	DictNeteaseFE string
	//go:embed wordlists/色情类型.txt
	DictSexual string
	//go:embed wordlists/色情词库.txt
	DictPornography string
	//go:embed wordlists/补充词库.txt
	DictAdditional string
	//go:embed wordlists/贪腐词库.txt
	DictCorruption string
	//go:embed wordlists/零时-Tencent.txt
	DictTemporaryTencent string
	//go:embed wordlists/非法网址.txt
	DictIllegalURL string
)
