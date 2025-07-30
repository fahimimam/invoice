package cmd

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/keepalive"
	"os"
)

func newGrpcConnection() (*grpc.ClientConn, error) {
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

	return connection, nil
}

func newGatewayFromIdentity(grpcConn *grpc.ClientConn, certPEM, keyPEM, mspID string) (*client.Gateway, *client.Contract, error) {

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

	channelName := envOrDefault("CHANNEL_NAME", "mychannel")
	chaincodeName := envOrDefault("CHAINCODE_NAME", "identity")

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return gw, contract, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func envOrDefault(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
