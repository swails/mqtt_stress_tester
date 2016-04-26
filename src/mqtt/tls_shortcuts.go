package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// Creates an anonymous TLS configuration from a certificate authority with no
// client certificates needed
func NewTLSAnonymousConfig(ca string) (*tls.Config, error) {

	caCert, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, fmt.Errorf("error reading CA file: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("error appending CA to cert pool")
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ClientAuth: tls.NoClientCert,
		RootCAs:    caCertPool,
	}

	return cfg, nil
}
