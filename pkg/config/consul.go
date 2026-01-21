package config

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// ConsulConfigCenter implements ConfigCenter using Consul
type ConsulConfigCenter struct {
	client *viper.Viper
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewConsulConfigCenter creates a new Consul configuration center
func NewConsulConfigCenter(opts *ConfigCenterOptions) (*ConsulConfigCenter, error) {
	log.Debugf("NewConsulConfigCenter %#v", opts)
	v := viper.New()
	v.AddRemoteProvider("consul", opts.Address, opts.Path)
	v.SetConfigType(opts.ContentType)
	// Read from Consul
	err := v.ReadRemoteConfig()
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})

	c := &ConsulConfigCenter{
		client: v,
		stopCh: stopCh,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic: %v\n\n%s", r, string(debug.Stack()))
			}
		}()

		for {
			select {
			case <-stopCh:
				log.Info("ConsulConfigCenter watcher stopped")
				return
			case <-time.After(time.Second * 5):
				err := v.WatchRemoteConfig()
				if err != nil {
					// Suppress harmless "No Files Found" error which can happen if key exists but watcher is flaky
					if err.Error() == "Remote Configurations Error: No Files Found" {
						log.Warnf("WatchRemoteConfig: %v (Retrying...)", err)
						continue
					}
					log.Errorf("WatchRemoteConfig error: %v", err)
				}
			}
		}
	}()

	return c, nil
}

// GetKey retrieves a configuration value from Consul
func (c *ConsulConfigCenter) GetKey(key string) any {
	return c.client.Get(key)
}

// Close closes the Consul client connection
func (c *ConsulConfigCenter) Close() error {
	close(c.stopCh)
	c.wg.Wait()
	return nil
}

// GetClient returns the underlying viper client
func (c *ConsulConfigCenter) GetClient() *viper.Viper {
	return c.client
}
