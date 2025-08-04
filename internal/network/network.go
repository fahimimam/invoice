package network

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type NetworkInterface interface {
	NetworkConnectionClose() error
	SetIdentityGateway(certPEM, keyPEM, mspID string) error
	CloseIdentityGateway() error
	SubmitTransaction(name string, args ...string) error
	EvaluateTransaction(name string, args ...string) ([]byte, error)
}

type Network struct {
	grpcConn      *grpc.ClientConn
	chaincodeName string
	channelName   string

	smartContract   *client.Contract
	identityGateway *client.Gateway
}

func NewNetworkRPCConnection() (*Network, error) {
	tlsCertPath := envOrDefault("TLS_CERT_PATH", "/etc/secret-volume/tlsCertPath")
	peerEndpoint := envOrDefault("PEER_ENDPOINT", "test-network-org1-peer1-peer.localho.st:443")
	gatewayPeer := envOrDefault("GATEWAY_PEER", "test-network-org1-peer1-peer.localho.st")

	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	channelName := envOrDefault("CHANNEL_NAME", "mychannel")
	chaincodeName := envOrDefault("CHAINCODE_NAME", "identity")

	return &Network{
		grpcConn:      connection,
		chaincodeName: chaincodeName,
		channelName:   channelName,
	}, nil
}

func (n *Network) NetworkConnectionClose() error {
	return n.grpcConn.Close()
}

func (n *Network) SetIdentityGateway(certPEM, keyPEM, mspID string) error {
	certBytes, err := base64.StdEncoding.DecodeString(certPEM)
	if err != nil {
		return err
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyPEM)
	if err != nil {
		return err
	}

	certificate, err := identity.CertificateFromPEM(certBytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(n.grpcConn))
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	n.smartContract = gw.GetNetwork(n.channelName).GetContract(n.chaincodeName)
	n.identityGateway = gw

	return nil
}

func (n *Network) CloseIdentityGateway() error {
	return n.identityGateway.Close()
}

func (n *Network) SubmitTransaction(name string, args ...string) error {
	_, err := n.smartContract.SubmitTransaction(name, args...)
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return n.smartContract.EvaluateTransaction(name, args...)
}

func envOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}
