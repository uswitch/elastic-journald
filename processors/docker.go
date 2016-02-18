package processors

import (
	"github.com/docker/engine-api/client"
	"github.com/golang/groupcache/lru"
)

type DockerEnricher struct {
	Client *client.Client
	Cache  *lru.Cache
}

func NewDockerEnricher() *DockerEnricher {
	defaultHeaders := map[string]string{"User-Agent": "elastic-journald"}
	client, err := client.NewClient("unix:///var/run/docker.sock", "v1.21", nil, defaultHeaders)

	if err != nil {
		panic("docker")
	}

	enricher := &DockerEnricher{
		Client: client,
		Cache:  lru.New(1024),
	}
	return enricher
}

func (d *DockerEnricher) lookupImage(containerId string) string {
	if image, ok := d.Cache.Get(containerId); ok {
		return image.(string)
	}

	container, err := d.Client.ContainerInspect(containerId)
	if err != nil {
		panic("docker")
	}

	image := container.Config.Image
	d.Cache.Add(containerId, image)
	return image
}

func (d *DockerEnricher) Process(entry LogEntry) {
	if containerId, ok := entry["container_id_full"]; ok {
		imageName := d.lookupImage(containerId.(string)) // TODO: cast nicely
		entry["docker_image"] = imageName
	}
}
