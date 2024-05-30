package duration_cache

import (
	"encoding/json"
	"sync"

	"github.com/coocood/freecache"
	"golang.org/x/sync/singleflight"
)

var gCache *freecache.Cache
var size = 100 * 1024 * 1024 //默认大小100M
var once = sync.Once{}
var sf singleflight.Group

const maxSize = 1024 * 1024 * 1024 //1个G

// 如果需要修改cache大小，需要在模块启动时，调用Get前调用此方法
func SetSize(sz int) {
	if sz > maxSize {
		sz = maxSize
	}
	size = sz
}

func GetWithBuild[T any](key string, f func() (T, error), duration int, rs *T, enFn func(T) ([]byte, error), deFn func([]byte, *T) error) error {
	initCache()
	err := gCache.GetFn([]byte(key), func(bytes []byte) error {
		return deFn(bytes, rs)
	})
	if err != nil {
		//防止本地并发set cache，加一个singleflight
		r, err, _ := sf.Do(key, func() (interface{}, error) {
			r, err := f()
			if err != nil {
				return nil, err
			}
			bt, err := enFn(r)
			if err != nil {
				return nil, err
			}
			err = gCache.Set([]byte(key), bt, duration)
			if err != nil {
				return nil, err
			}
			return r, nil
		})
		if err != nil {
			return err
		}
		*rs = r.(T)
	}
	return nil
}

/*
*
用于缓存一些需要rpc获得结果的值，一般用来保存setting值
key为缓存区别的唯一值
f 为获取实际数据的方法，比如getSetting
duration为想要缓存的时长，这个时长为最长时长，可能受到全局cache LRU之类算法影响
rs 是最终的返回值，需要是T的指针类型
*/
func Get[T any](key string, f func() (T, error), duration int, rs *T) error {
	return GetWithBuild(key, f, duration, rs, func(t T) ([]byte, error) {
		return makeBytes(t)
	}, func(bytes []byte, t *T) error {
		return makeFromBytes(bytes, t)
	})
}

func initCache() {
	once.Do(func() {
		gCache = freecache.NewCache(size)
	})
}

func makeBytes[T any](v T) ([]byte, error) {
	return json.Marshal(v)
}

func makeFromBytes[T any](bt []byte, rs T) error {
	return json.Unmarshal(bt, rs)
}
