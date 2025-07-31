package gateways

import (
	"encoding/base64"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/triapex/auth/config"
	"google.golang.org/grpc"
)

func NewGatewayForIdentity(grpcConn *grpc.ClientConn, certPEM, keyPEM, mspID string, identityConfig *config.Identity) (*client.Gateway, *client.Contract, error) {

	certBytes, err := base64.StdEncoding.DecodeString(certPEM)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyPEM)
	if err != nil {
		return nil, nil, err
	}

	certificate, err := identity.CertificateFromPEM(certBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create identity: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create signer: %w", err)
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(grpcConn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	channelName := identityConfig.Channel
	chaincodeName := identityConfig.Channel

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return gw, contract, nil
}
