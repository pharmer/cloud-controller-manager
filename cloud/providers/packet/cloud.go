package packet

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/packethost/packngo"
	cloudprovider "k8s.io/cloud-provider"
)

const (
	ProviderName = "packet"
)

type credential struct {
	Project string `json:"project" yaml:"project"`
	ApiKey  string `json:"apiKey" yaml:"apiKey"`
	Zone    string `json:"zone" yaml:"zone"`
}

type Cloud struct {
	client        *packngo.Client
	instances     cloudprovider.Instances
	zones         cloudprovider.Zones
	loadbalancers cloudprovider.LoadBalancer
}

func init() {
	cloudprovider.RegisterCloudProvider(
		ProviderName,
		func(config io.Reader) (cloudprovider.Interface, error) {
			return newCloud(config)
		})
}

func newCloud(config io.Reader) (*Cloud, error) {
	packet := &credential{}
	contents, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, packet)
	if err != nil {
		return nil, err
	}

	packetClient := packngo.NewClientWithAuth("", packet.ApiKey, nil)

	return &Cloud{
		client:        packetClient,
		instances:     newInstances(packetClient, packet.Project),
		zones:         newZones(packetClient, packet.Project, packet.Zone),
		loadbalancers: newLoadbalancers(packetClient),
	}, nil
}

func (c *Cloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
}

func (c *Cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return c.loadbalancers, true
}

func (c *Cloud) Instances() (cloudprovider.Instances, bool) {
	return c.instances, true
}

func (c *Cloud) Zones() (cloudprovider.Zones, bool) {
	return c.zones, true
}

func (c *Cloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (c *Cloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (c *Cloud) ProviderName() string {
	return ProviderName
}

func (c *Cloud) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nil, nil
}

func (c *Cloud) HasClusterID() bool {
	return true
}
