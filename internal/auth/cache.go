package auth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/pkg/errors"
)

const dirName = "jetpack"

func (a *Authenticator) fetchJWKSWithCache() (*keyfunc.JWKS, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", a.Domain)
	fileName := fmt.Sprintf("%s.jwks.json", a.Domain)
	baseDir, err := os.UserCacheDir()
	if err != nil {
		baseDir = "~/.cache"
	}
	// example ~/.cache/jetpack/accounts.jetpack.io.jwks.json
	cacheDir := filepath.Join(baseDir, dirName)
	path := filepath.Join(cacheDir, fileName)

	// check Cache if miss, jwksJSON will be empty
	jwksJSON, err := readJWKSCache(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if jwksJSON == nil { // cache miss
		// save new keys to cache
		jwksJSON, err = saveJWKSCache(jwksURL, cacheDir, path)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	jwks, err := keyfunc.NewJSON(jwksJSON)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return jwks, nil
}

func readJWKSCache(path string) ([]byte, error) {
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	modificationTime := fileInfo.ModTime()
	current := time.Now()
	// cache duration: 1 hour
	if current.After(modificationTime.Add(time.Hour)) {
		return nil, nil
	}
	return os.ReadFile(path)
}

func saveJWKSCache(url string, cacheDir string, path string) ([]byte, error) {
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	// make sure cache dir exists before creating the file
	if err = os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, errors.WithStack(err)
	}
	out, err := os.Create(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer out.Close()
	jwks, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = out.Write(jwks)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return jwks, nil
}
