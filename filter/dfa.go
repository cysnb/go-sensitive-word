package filter

import "sync"

type dfaNode struct {
	children map[rune]*dfaNode
	isLeaf   bool
}

func newDfaNode() *dfaNode {
	return &dfaNode{children: make(map[rune]*dfaNode)}
}

// DfaModel 基于 DFA 的敏感词匹配器。
type DfaModel struct {
	mu       sync.RWMutex
	root     *dfaNode
	stopCh   chan struct{}
	stopOnce sync.Once
}

// DFAModel 是 DfaModel 的兼容别名。
type DFAModel = DfaModel

// NewDfaModel 创建 DFA 匹配器。
func NewDfaModel() *DfaModel {
	return &DfaModel{
		root:   newDfaNode(),
		stopCh: make(chan struct{}),
	}
}

// NewDFAModel 创建 DFA 匹配器。
func NewDFAModel() *DfaModel {
	return NewDfaModel()
}

// AddWords 添加多个词。
func (m *DfaModel) AddWords(words ...string) {
	for _, word := range words {
		m.AddWord(word)
	}
}

// AddWord 添加单个词。
func (m *DfaModel) AddWord(word string) {
	if word == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.root
	for _, r := range word {
		next, ok := now.children[r]
		if !ok {
			next = newDfaNode()
			now.children[r] = next
		}
		now = next
	}
	now.isLeaf = true
}

// DelWords 删除多个词。
func (m *DfaModel) DelWords(words ...string) {
	for _, word := range words {
		m.DelWord(word)
	}
}

// DelWord 删除单个词。
func (m *DfaModel) DelWord(word string) {
	if word == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	runes := []rune(word)
	type pathElem struct {
		node *dfaNode
		ch   rune
	}
	path := make([]pathElem, 0, len(runes)+1)
	now := m.root
	path = append(path, pathElem{node: now})
	for _, r := range runes {
		next, ok := now.children[r]
		if !ok {
			return
		}
		path = append(path, pathElem{node: next, ch: r})
		now = next
	}
	if !now.isLeaf {
		return
	}
	now.isLeaf = false
	for i := len(path) - 1; i >= 1; i-- {
		curr := path[i].node
		parent := path[i-1].node
		ch := path[i].ch
		if len(curr.children) == 0 && !curr.isLeaf {
			delete(parent.children, ch)
			continue
		}
		break
	}
}

// Listen 监听词库变更，applied 在每条变更应用后调用。
func (m *DfaModel) Listen(addChan, delChan <-chan string, applied func()) {
	go func() {
		for {
			select {
			case word, ok := <-addChan:
				if !ok {
					return
				}
				m.AddWord(word)
				if applied != nil {
					applied()
				}
			case <-m.stopCh:
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case word, ok := <-delChan:
				if !ok {
					return
				}
				m.DelWord(word)
				if applied != nil {
					applied()
				}
			case <-m.stopCh:
				return
			}
		}
	}()
}

// Close 停止监听。
func (m *DfaModel) Close() error {
	m.stopOnce.Do(func() { close(m.stopCh) })
	return nil
}

func (m *DfaModel) findOneLocked(text string) string {
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found || (!now.isLeaf && pos == length-1) {
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf && start <= pos {
			return string(runes[start : pos+1])
		}
		parent = now
	}
	return ""
}

// FindOne 查找第一个敏感词。
func (m *DfaModel) FindOne(text string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.findOneLocked(text)
}

// IsSensitive 判断是否存在敏感词。
func (m *DfaModel) IsSensitive(text string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.findOneLocked(text) != ""
}

// FindAll 查找所有敏感词（去重）。
func (m *DfaModel) FindAll(text string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var matches []string
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found {
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf && start <= pos {
			matches = append(matches, string(runes[start:pos+1]))
		}
		if pos == length-1 {
			parent = m.root
			pos = start
			start++
			continue
		}
		parent = now
	}
	res := make([]string, 0, len(matches))
	set := make(map[string]struct{}, len(matches))
	for _, word := range matches {
		if _, ok := set[word]; !ok {
			set[word] = struct{}{}
			res = append(res, word)
		}
	}
	return res
}

// FindAllCount 查找所有敏感词及出现次数。
func (m *DfaModel) FindAllCount(text string) map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make(map[string]int)
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found {
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf && start <= pos {
			res[string(runes[start:pos+1])]++
		}
		if pos == length-1 {
			parent = m.root
			pos = start
			start++
			continue
		}
		parent = now
	}
	return res
}

// FindAllRanges 返回所有匹配区间。
func (m *DfaModel) FindAllRanges(text string) []Range {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ranges []Range
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	seen := make(map[Range]struct{})
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found {
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf && start <= pos {
			r := Range{Start: start, End: pos}
			if _, exists := seen[r]; !exists {
				seen[r] = struct{}{}
				ranges = append(ranges, r)
			}
		}
		if pos == length-1 {
			parent = m.root
			pos = start
			start++
			continue
		}
		parent = now
	}
	return ranges
}

// Replace 将敏感词替换为指定字符。
func (m *DfaModel) Replace(text string, repl rune) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found || (!now.isLeaf && pos == length-1) {
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf && start <= pos {
			for i := start; i <= pos; i++ {
				runes[i] = repl
			}
		}
		parent = now
	}
	return string(runes)
}

// Remove 删除所有敏感词。
func (m *DfaModel) Remove(text string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var found bool
	var now *dfaNode
	start := 0
	parent := m.root
	runes := []rune(text)
	length := len(runes)
	filtered := make([]rune, 0, length)
	for pos := 0; pos < length; pos++ {
		now, found = parent.children[runes[pos]]
		if !found || (!now.isLeaf && pos == length-1) {
			filtered = append(filtered, runes[start])
			parent = m.root
			pos = start
			start++
			continue
		}
		if now.isLeaf {
			start = pos + 1
			parent = m.root
		} else {
			parent = now
		}
	}
	filtered = append(filtered, runes[start:]...)
	return string(filtered)
}

var _ Filter = (*DfaModel)(nil)
var _ RangedFilter = (*DfaModel)(nil)
