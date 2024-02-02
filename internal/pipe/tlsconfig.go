package pipe

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	tls "github.com/refraction-networking/utls"
)

// createServerTLSConfig creates a TLS pipe configuration with the specified
// pipe name. It returns a *tls.Config that will be used by the pipe.
func createServerTLSConfig(tlsServerName string) (tlsConfig *tls.Config) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("cannot generate RSA key: %v", err))
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		panic(fmt.Sprintf("failed to generate serial number: %v", err))
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(5 * 365 * time.Hour * 24)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{tlsServerName},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	template.DNSNames = append(template.DNSNames, tlsServerName)

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to create certificate: %v", err))
	}

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		panic(fmt.Sprintf("failed to create certificate: %v", err))
	}

	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(certPem)

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   tlsServerName,
		RootCAs:      roots,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig
}

func publicKey(priv any) (pub any) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}
