package store

type NotifyCache interface {
	LRUExists(key string) bool
	LRUAdd(key string)
}