/**
 * @Author: zjj
 * @Date: 2024/4/10
 * @Desc:
**/

package wordmonitor

import (
	"crypto/md5"
	"encoding/base64"
)

const (
	maxLength = 99999
)

type wordCache struct {
	Set map[string]struct{}
}

func newWordCache() *wordCache {
	return &wordCache{
		Set: make(map[string]struct{}),
	}
}

func (c wordCache) Add(key string) {
	if len(c.Set) > maxLength {
		c.Set = make(map[string]struct{})
	}
	b := md5.Sum([]byte(key))
	k := base64.StdEncoding.EncodeToString(b[:])
	c.Set[k] = struct{}{}
}

func (c wordCache) Exit(key string) bool {
	b := md5.Sum([]byte(key))
	k := base64.StdEncoding.EncodeToString(b[:])
	_, ok := c.Set[k]
	return ok
}
