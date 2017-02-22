package collector

import "github.com/fsouza/go-dockerclient"

// Stats represents singe stat from docker stats api for specific task
type Stats struct {
	Tags  map[string]string
	Stats docker.Stats
}
