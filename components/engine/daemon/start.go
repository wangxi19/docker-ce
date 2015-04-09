package daemon

import (
	"fmt"

	"github.com/docker/docker/engine"
	"github.com/docker/docker/runconfig"
)

func (daemon *Daemon) ContainerStart(name string, env *engine.Env) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}

	if container.IsPaused() {
		return fmt.Errorf("Cannot start a paused container, try unpause instead.")
	}

	if container.IsRunning() {
		return fmt.Errorf("Container already started")
	}

	// If no environment was set, then no hostconfig was passed.
	// This is kept for backward compatibility - hostconfig should be passed when
	// creating a container, not during start.
	if len(env.Map()) > 0 {
		hostConfig := runconfig.ContainerHostConfigFromJob(env)
		if err := daemon.setHostConfig(container, hostConfig); err != nil {
			return err
		}
	}
	if err := container.Start(); err != nil {
		container.LogEvent("die")
		return fmt.Errorf("Cannot start container %s: %s", name, err)
	}

	return nil
}

func (daemon *Daemon) setHostConfig(container *Container, hostConfig *runconfig.HostConfig) error {
	container.Lock()
	defer container.Unlock()
	if err := parseSecurityOpt(container, hostConfig); err != nil {
		return err
	}

	// Register any links from the host config before starting the container
	if err := daemon.RegisterLinks(container, hostConfig); err != nil {
		return err
	}

	container.hostConfig = hostConfig
	container.toDisk()

	return nil
}
