package sec

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
)

type Certifier struct {
	host            string
	org             string
	validFrom       time.Time
	validFor        time.Duration
	isCA            bool
	rsaBits         int
	certOut, keyOut string
}

func (c Certifier) NewCertificate() (err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate elliptic p256: %w", err)
	}
	keyUsage := x509.KeyUsageDigitalSignature

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{c.org},
		},
		NotBefore:             c.validFrom,
		NotAfter:              c.validFrom.Add(c.validFor),
		KeyUsage:              keyUsage | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	hosts := strings.Split(c.host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	certOut, err := os.Create(c.certOut)
	if err != nil {
		return fmt.Errorf("create cert out %q: %w", c.certOut, err)
	}
	defer func() {
		err = dcheck.Wrap(certOut.Close(), err, "close %q", c.certOut)
	}()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return fmt.Errorf("pem encode cert out: %w", err)
	}

	slog.Debug("wrote pem", "to", c.certOut)

	keyOut, err := os.OpenFile(c.keyOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open key out %q: %w", c.keyOut, err)
	}
	defer func() {
		err = dcheck.Wrap(keyOut.Close(), err, "close %q", c.keyOut)
	}()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return fmt.Errorf("pem encode key out: %w", err)
	}

	slog.Debug("wrote key", "to", c.keyOut)

	return nil
}

type CertifierOption func(Certifier) Certifier

func DefaultCertifier(host string) Certifier {
	return Certifier{
		certOut:   "cert.pem",
		keyOut:    "cert.key",
		host:      host,
		org:       "Co",
		validFrom: time.Now(),
		validFor:  13 * 24 * time.Hour,
		isCA:      true,
		rsaBits:   3072,
	}
}

func New(host string, opts ...CertifierOption) Certifier {
	c := DefaultCertifier(host)
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}
