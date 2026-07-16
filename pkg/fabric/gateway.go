package fabric

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type FabricGateway struct {
	Contract *client.Contract
}

func NewGateway() (*FabricGateway, error) {
	cryptoPath := os.Getenv("CRYPTO_PATH")

	certPath := filepath.Join(cryptoPath,
		"users/Admin@org1.example.com/msp/signcerts/cert.pem")

	keyDir := filepath.Join(cryptoPath,
		"users/Admin@org1.example.com/msp/keystore")

	tlsPath := filepath.Join(cryptoPath,
		"peers/peer0.org1.example.com/tls/ca.crt")

	tlsCert, err := os.ReadFile(tlsPath)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(tlsCert) {
		return nil, fmt.Errorf("failed to add TLS cert")
	}

	creds := credentials.NewClientTLSFromCert(pool, "peer0.org1.example.com")

	conn, err := grpc.NewClient("localhost:7051", grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	idCertBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	cert, err := identity.CertificateFromPEM(idCertBytes)
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity("Org1MSP", cert)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("keystore is empty")
	}

	keyPath := filepath.Join(keyDir, files[0].Name())

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
	)
	if err != nil {
		return nil, err
	}

	network := gw.GetNetwork("mychannel")
	contract := network.GetContract("document") 

	return &FabricGateway{
		Contract: contract,
	}, nil
}