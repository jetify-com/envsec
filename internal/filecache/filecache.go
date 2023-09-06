// filecache is a simple local file-based cache
package filecache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.jetpack.io/envsec/internal/xdg"
)

var NotFound = errors.New("not found")
var Expired = errors.New("expired")

const prefix = "filecache-"

type cache struct {
	appName string
}

func New(appName string) *cache {
	return &cache{appName: appName}
}

type data struct {
	Val []byte
	Exp time.Time
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
	dir := xdg.CacheSubpath(c.appName)
	_ = os.MkdirAll(dir, 0755)
	return xdg.CacheSubpath(filepath.Join(c.appName, prefix+key))
}
