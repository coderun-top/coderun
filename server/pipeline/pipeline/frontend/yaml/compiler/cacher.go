package compiler

import (
	"path"
	"strings"

	"gitlab.com/douwa/registry/dougo/src/dougo/core/pipeline/pipeline/frontend/yaml"

	libcompose "github.com/docker/libcompose/yaml"
)

// Cacher defines a compiler transform that can be used
// to implement default caching for a repository.
type Cacher interface {
	Restore(repo string, cache *yaml.Cache) *yaml.Container
	Save(repo string, cache *yaml.Cache) *yaml.Container
}

type volumeCacher struct {
	base string
}

func (c *volumeCacher) Restore(repo string, cache *yaml.Cache) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "crun/volume-cache",
		Vargs: map[string]interface{}{
			"mount":       cache.Paths,
			"path":        "/cache",
			"restore":     true,
			"file":        strings.Replace(cache.Key, "/", "_", -1) + ".tar",
			"fallback_to": "master.tar",
		},
		Volumes: libcompose.Volumes{
			Volumes: []*libcompose.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

func (c *volumeCacher) Save(repo string, cache *yaml.Cache) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "crun/volume-cache",
		Vargs: map[string]interface{}{
			"mount":   cache.Paths,
			"path":    "/cache",
			"rebuild": true,
			"flush":   true,
			"file":    strings.Replace(cache.Key, "/", "_", -1) + ".tar",
		},
		Volumes: libcompose.Volumes{
			Volumes: []*libcompose.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

type s3Cacher struct {
	bucket string
	access string
	secret string
	region string
}

func (c *s3Cacher) Restore(repo string, cache *yaml.Cache) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Vargs: map[string]interface{}{
			"mount":      cache.Paths,
			"access_key": c.access,
			"secret_key": c.secret,
			"bucket":     c.bucket,
			"region":     c.region,
			"rebuild":    true,
		},
	}
}

func (c *s3Cacher) Save(repo string, cache *yaml.Cache) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Vargs: map[string]interface{}{
			"mount":      cache.Paths,
			"access_key": c.access,
			"secret_key": c.secret,
			"bucket":     c.bucket,
			"region":     c.region,
			"rebuild":    true,
			"flush":      true,
		},
	}
}
