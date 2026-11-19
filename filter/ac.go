package filter

import (
	"sync"
	"sync/atomic"
	"time"
)

type acNode struct {
	children map[rune]*acNode
	fail     *acNode
	output   []string
}

func newAcNode() *acNode {
	return &acNode{children: make(map[rune]*acNode), output: make([]string, 0)}
}

// AcModel 基于 AC 自动机的敏感词匹配器。
type AcModel struct {
	rootPtr     atomic.Pointer[acNode]
	mu          sync.Mutex
	pendingAdds []string
	pendingDels []string
	ticker      *time.Ticker
	done        chan struct{}
	closeOnce   sync.Once
}

// ACModel 是 AcModel 的兼容别名。
type ACModel = AcModel

// NewAcModel 创建 AC 匹配器。
func NewAcModel() *AcModel {
	model := &AcModel{done: make(chan struct{})}
	model.rootPtr.Store(newAcNode())
	return model
}

// NewACModel 创建 AC 匹配器。
func NewACModel() *AcModel {
	return NewAcModel()
}

// AddWord 添加单个词。
func (m *AcModel) AddWord(word string) {
	if word == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	root := m.rootPtr.Load()
	newRoot := m.cloneNode(root)
	now := newRoot
	for _, r := range word {
		next, ok := now.children[r]
		if !ok {
			next = newAcNode()
			now.children[r] = next
		}
		now = next
	}
	now.output = append(now.output, word)
	m.buildFailurePointer(newRoot)
	m.rootPtr.Store(newRoot)
}

// AddWords 添加多个词。
func (m *AcModel) AddWords(words ...string) {
	for _, word := range words {
		m.AddWord(word)
	}
}

// DelWord 删除单个词。
func (m *AcModel) DelWord(word string) {
	if word == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	root := m.rootPtr.Load()
	newRoot := m.cloneNode(root)
	now := newRoot
	for _, r := range word {
		next, ok := now.children[r]
		if !ok {
			return
		}
		now = next
	}
	for i, w := range now.output {
		if w == word {
			now.output = append(now.output[:i], now.output[i+1:]...)
			break
		}
	}
	m.buildFailurePointer(newRoot)
	m.rootPtr.Store(newRoot)
}

// DelWords 删除多个词。
func (m *AcModel) DelWords(words ...string) {
	for _, word := range words {
		m.DelWord(word)
	}
}

func (m *AcModel) cloneNode(n *acNode) *acNode {
	newNode := &acNode{
		children: make(map[rune]*acNode, len(n.children)),
		output:   make([]string, len(n.output)),
	}
	copy(newNode.output, n.output)
	for r, child := range n.children {
		newNode.children[r] = m.cloneNode(child)
	}
	return newNode
}

func (m *AcModel) buildFailurePointer(root *acNode) {
	queue := make([]*acNode, 0)
	for _, child := range root.children {
		child.fail = root
		queue = append(queue, child)
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for char, child := range current.children {
			queue = append(queue, child)
			temp := current.fail
			for temp != nil && temp.children[char] == nil {
				temp = temp.fail
			}
			if temp == nil {
				child.fail = root
			} else {
				child.fail = temp.children[char]
				child.output = append(child.output, child.fail.output...)
			}
		}
	}
}

// Listen 启动监听与窗口合并。
func (m *AcModel) Listen(addChan, delChan <-chan string, applied func()) {
	m.mu.Lock()
	m.pendingAdds = make([]string, 0, 1000)
	m.pendingDels = make([]string, 0, 1000)
	m.ticker = time.NewTicker(100 * time.Millisecond)
	ticker := m.ticker
	m.mu.Unlock()

	go func() {
		defer ticker.Stop()
		for {
			select {
			case word, ok := <-addChan:
				if !ok {
					return
				}
				if word == "" {
					continue
				}
				m.mu.Lock()
				m.pendingAdds = append(m.pendingAdds, word)
				shouldFlush := len(m.pendingAdds) >= 1000
				m.mu.Unlock()
				if shouldFlush {
					m.flushPending(applied)
				}
			case word, ok := <-delChan:
				if !ok {
					return
				}
				if word == "" {
					continue
				}
				m.mu.Lock()
				m.pendingDels = append(m.pendingDels, word)
				shouldFlush := len(m.pendingDels) >= 1000
				m.mu.Unlock()
				if shouldFlush {
					m.flushPending(applied)
				}
			case <-ticker.C:
				m.flushPending(applied)
			case <-m.done:
				m.flushPending(applied)
				return
			}
		}
	}()
}

func (m *AcModel) flushPending(applied func()) {
	m.mu.Lock()
	if len(m.pendingAdds) == 0 && len(m.pendingDels) == 0 {
		m.mu.Unlock()
		return
	}
	adds := m.pendingAdds
	dels := m.pendingDels
	m.pendingAdds = make([]string, 0, 1000)
	m.pendingDels = make([]string, 0, 1000)
	m.mu.Unlock()

	root := m.rootPtr.Load()
	newRoot := m.cloneNode(root)

	for _, word := range dels {
		now := newRoot
		ok := true
		for _, r := range word {
			next, exists := now.children[r]
			if !exists {
				ok = false
				break
			}
			now = next
		}
		if !ok {
			continue
		}
		for i, w := range now.output {
			if w == word {
				now.output = append(now.output[:i], now.output[i+1:]...)
				break
			}
		}
	}

	for _, word := range adds {
		now := newRoot
		for _, r := range word {
			next, ok := now.children[r]
			if !ok {
				next = newAcNode()
				now.children[r] = next
			}
			now = next
		}
		now.output = append(now.output, word)
	}

	m.buildFailurePointer(newRoot)
	m.rootPtr.Store(newRoot)

	if applied != nil {
		for range adds {
			applied()
		}
		for range dels {
			applied()
		}
	}
}

// Close 停止监听。
func (m *AcModel) Close() error {
	m.closeOnce.Do(func() { close(m.done) })
	return nil
}

// FindAll 查找所有敏感词（去重）。
func (m *AcModel) FindAll(text string) []string {
	root := m.rootPtr.Load()
	var matches []string
	seen := make(map[string]struct{})
	now := root
	for _, r := range text {
		for now != root && now.children[r] == nil {
			now = now.fail
		}
		if next, ok := now.children[r]; ok {
			now = next
		} else {
			now = root
		}
		for _, w := range now.output {
			if _, ok := seen[w]; !ok {
				seen[w] = struct{}{}
				matches = append(matches, w)
			}
		}
	}
	return matches
}

// FindAllCount 查找所有敏感词及出现次数。
func (m *AcModel) FindAllCount(text string) map[string]int {
	root := m.rootPtr.Load()
	counts := make(map[string]int)
	now := root
	for _, r := range text {
		for now != root && now.children[r] == nil {
			now = now.fail
		}
		if next, ok := now.children[r]; ok {
			now = next
		} else {
			now = root
		}
		for _, w := range now.output {
			counts[w]++
		}
	}
	return counts
}

// FindOne 查找第一个敏感词（最长匹配）。
func (m *AcModel) FindOne(text string) string {
	root := m.rootPtr.Load()
	now := root
	for _, r := range text {
		for now != root && now.children[r] == nil {
			now = now.fail
		}
		if next, ok := now.children[r]; ok {
			now = next
		} else {
			now = root
		}
		if len(now.output) > 0 {
			longest := now.output[0]
			for i := 1; i < len(now.output); i++ {
				if len([]rune(now.output[i])) > len([]rune(longest)) {
					longest = now.output[i]
				}
			}
			return longest
		}
	}
	return ""
}

// IsSensitive 判断是否存在敏感词。
func (m *AcModel) IsSensitive(text string) bool {
	return m.FindOne(text) != ""
}

// FindAllRanges 返回所有匹配区间。
func (m *AcModel) FindAllRanges(text string) []Range {
	root := m.rootPtr.Load()
	var ranges []Range
	seen := make(map[Range]struct{})
	now := root
	runes := []rune(text)
	for i, r := range runes {
		for now != root && now.children[r] == nil {
			now = now.fail
		}
		if next, ok := now.children[r]; ok {
			now = next
		} else {
			now = root
		}
		for _, w := range now.output {
			wordLen := len([]rune(w))
			start := i - wordLen + 1
			if start >= 0 {
				r := Range{Start: start, End: i}
				if _, ok := seen[r]; !ok {
					seen[r] = struct{}{}
					ranges = append(ranges, r)
				}
			}
		}
	}
	return ranges
}

// Replace 将敏感词替换为指定字符。
func (m *AcModel) Replace(text string, repl rune) string {
	ranges := m.FindAllRanges(text)
	runes := []rune(text)
	marked := make([]bool, len(runes))
	for _, r := range ranges {
		for i := r.Start; i <= r.End && i < len(marked); i++ {
			marked[i] = true
		}
	}
	for i := range marked {
		if marked[i] {
			runes[i] = repl
		}
	}
	return string(runes)
}

// Remove 删除所有敏感词。
func (m *AcModel) Remove(text string) string {
	ranges := m.FindAllRanges(text)
	runes := []rune(text)
	toDelete := make([]bool, len(runes))
	for _, r := range ranges {
		for i := r.Start; i <= r.End && i < len(toDelete); i++ {
			toDelete[i] = true
		}
	}
	result := make([]rune, 0, len(runes))
	for i, r := range runes {
		if !toDelete[i] {
			result = append(result, r)
		}
	}
	return string(result)
}

var _ Filter = (*AcModel)(nil)
var _ RangedFilter = (*AcModel)(nil)
