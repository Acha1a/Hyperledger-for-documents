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

func InitFabric() (*client.Contract, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	// 2. Формируем АБСОЛЮТНЫЙ путь к папке организации
	basePath := filepath.Join(home, "hyperledger/blockchain/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com")
	
	// 3. Уточняем пути к ключам и сертификатам
	mspPath := filepath.Join(basePath, "users", "User1@org1.example.com", "msp")
	certPath := filepath.Join(mspPath, "signcerts", "cert.pem")
	keyPath := filepath.Join(mspPath, "keystore")

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %w", err)
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	id, err := identity.NewX509Identity("Org1MSP", cert)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(keyPath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("keystore is empty")
	}

	keyBytes, err := os.ReadFile(filepath.Join(keyPath, files[0].Name()))
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, err
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	tlsCertPath := filepath.Join(basePath, "peers", "peer0.org1.example.com", "tls", "ca.crt")
	creds, err := credentials.NewClientTLSFromFile(tlsCertPath, "peer0.org1.example.com")
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial("localhost:7051", grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(conn))
	if err != nil {
		return nil, err
	}

	return gw.GetNetwork("mychannel").GetContract("basic"), nil
}