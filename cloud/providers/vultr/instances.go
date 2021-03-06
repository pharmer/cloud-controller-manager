package vultr

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gv "github.com/JamesClonk/vultr/lib"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"pharmer.dev/cloud-controller-manager/cloud"
)

type instances struct {
	client *gv.Client
}

func newInstances(client *gv.Client) cloudprovider.Instances {
	return &instances{client}
}

func (i *instances) NodeAddresses(_ context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	server, err := serverByName(i.client, name)
	if err != nil {
		return nil, err
	}
	return nodeAddresses(server)
}

func (i *instances) NodeAddressesByProviderID(_ context.Context, providerID string) ([]v1.NodeAddress, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}
	server, err := serverByID(i.client, id)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(&server)
}

func nodeAddresses(server *gv.Server) ([]v1.NodeAddress, error) {
	var addresses []v1.NodeAddress
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeHostName, Address: server.Name})

	if server.InternalIP == "" {
		return nil, fmt.Errorf("could not get private ip")
	}
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: server.InternalIP})

	if server.MainIP == "" {
		return nil, fmt.Errorf("could not get public ip")
	}
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: server.MainIP})

	return addresses, nil
}

func (i *instances) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	return i.InstanceID(ctx, nodeName)
}

func (i *instances) InstanceID(_ context.Context, nodeName types.NodeName) (string, error) {
	server, err := serverByName(i.client, nodeName)
	if err != nil {
		return "", err
	}
	return server.ID, nil
}

func (i *instances) InstanceType(_ context.Context, nodeName types.NodeName) (string, error) {
	server, err := serverByName(i.client, nodeName)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(server.PlanID), nil
}

func (i *instances) InstanceTypeByProviderID(_ context.Context, providerID string) (string, error) {
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return "", err
	}
	server, err := serverByID(i.client, id)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(server.PlanID), nil
}

func (i *instances) AddSSHKeyToAllInstances(_ context.Context, user string, keyData []byte) error {
	return cloud.ErrNotImplemented
}

func (i *instances) CurrentNodeName(_ context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

func (i *instances) InstanceExistsByProviderID(_ context.Context, providerID string) (bool, error) {
	//TODO(sanjid): check provider id here
	id, err := serverIDFromProviderID(providerID)
	if err != nil {
		return false, err
	}
	_, err = serverByID(i.client, id)
	if err == nil {
		return true, nil
	}

	return false, nil
}

func (i *instances) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	return false, cloudprovider.NotImplemented
}

func serverByID(client *gv.Client, id string) (gv.Server, error) {
	return client.GetServer(id)
}

func serverByName(client *gv.Client, nodeName types.NodeName) (*gv.Server, error) {
	servers, err := client.GetServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Name == string(nodeName) {
			return &server, nil
		}
	}
	return nil, cloudprovider.InstanceNotFound
}

// serverIDFromProviderID returns a server's ID from providerID.
//
// The providerID spec should be retrievable from the Kubernetes
// node object. The expected format is: vultr://server-id

func serverIDFromProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", errors.New("providerID cannot be empty string")
	}

	split := strings.Split(providerID, "/")
	if len(split) != 3 {
		return "", fmt.Errorf("unexpected providerID format: %s, format should be: vultr://12345", providerID)
	}

	// since split[0] is actually "vultr:"
	if strings.TrimSuffix(split[0], ":") != ProviderName {
		return "", fmt.Errorf("provider name from providerID should be vultr: %s", providerID)
	}

	return split[2], nil
}
