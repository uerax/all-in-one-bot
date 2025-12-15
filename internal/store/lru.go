package store

import (
	"container/list"
	"sync"
)

type Value any

type Entry struct {
    Key   string
    Value Value
}

type LRU struct {
	cap int
	list *list.List
	cache map[string]*list.Element
	mu sync.Mutex
}

func NewLRU(capacity int) *LRU {
    if capacity <= 0 {
        capacity = 1 // 确保容量至少为 1
    }
    return &LRU{
        cap: 	  capacity,
        list:     list.New(),
        cache:    make(map[string]*list.Element, capacity),
    }
}

func (c *LRU) LRUExists(key string) bool {
    _, ok := c.Get(key)
    return ok
}
func (c *LRU) LRUAdd(key string) {
    c.Set(key, struct{}{})
}

func (c *LRU) Get(key string) (Value, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if element, ok := c.cache[key]; ok {
        // 1. 移动到头部：将该节点移到链表的最前端（Head）
        c.list.MoveToFront(element)
        
        // 2. 返回 Value
        return element.Value.(*Entry).Value, true
    }
    
    return nil, false
}

func (c *LRU) Set(key string, value Value) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // 1. 检查元素是否已存在
    if element, ok := c.cache[key]; ok {
        // A. 存在：更新值并移动到头部
        c.list.MoveToFront(element)
        element.Value.(*Entry).Value = value // 更新 Value
        return
    }

    // 2. 元素不存在：
    // B. 检查容量是否溢出
    if c.list.Len() >= c.cap {
        c.removeOldest() // 淘汰最冷元素
    }
    
    // C. 创建新 Entry 并添加到头部
    entry := &Entry{Key: key, Value: value}
    element := c.list.PushFront(entry)
    c.cache[key] = element
}

func (c *LRU) removeOldest() {
    // 获取尾部元素 (Tail)
    if tail := c.list.Back(); tail != nil {
        // 1. 从链表中移除节点
        c.list.Remove(tail)
        
        // 2. 从 Map 中删除对应的 Key
        // 从链表节点的值中取出 Key
        key := tail.Value.(*Entry).Key
        delete(c.cache, key)
    }
}
