// filecache is a simple local file-based cache
package filecache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

var NotFound = errors.New("not found")
var Expired = errors.New("expired")

type cache struct {
	domain   string
	cacheDir string
}

type data struct {
	Val []byte
	Exp time.Time
}

type Option func(*cache)

func New(domain string, opts ...Option) *cache {
	result := &cache{domain: domain}

	var err error
	result.cacheDir, err = os.UserCacheDir()
	if err != nil {
		result.cacheDir = "~/.cache"
	}

	for _, opt := range opts {
		opt(result)
	}

	return result
}

func WithCacheDir(dir string) Option {
	return func(c *cache) {
		c.cacheDir = dir
	}
}

func (c *cache) Set(key string, val []byte, dur time.Duration) error {
	d, err := json.Marshal(data{Val: val, Exp: time.Now().Add(dur)})
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(os.WriteFile(c.filename(key), d, 0644))
}

func (c *cache) SetT(key string, val []byte, t time.Time) error {
	d, err := json.Marshal(data{Val: val, Exp: t})
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(os.WriteFile(c.filename(key), d, 0644))
}

func (c *cache) Get(key string) ([]byte, error) {
	path := c.filename(key)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, NotFound
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	d := data{}
	if err := json.Unmarshal(content, &d); err != nil {
		return nil, errors.WithStack(err)
	}
	if time.Now().After(d.Exp) {
		return nil, Expired
	}
	return d.Val, nil
}

func (c *cache) filename(key string) string {
	dir := filepath.Join(c.cacheDir, c.domain)
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, key)
}
