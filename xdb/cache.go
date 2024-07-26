package xdb

import cache2 "github.com/daodao97/xgo/cache"

var cache cache2.Cache

func SetCache(c cache2.Cache) {
	cache = c
}
