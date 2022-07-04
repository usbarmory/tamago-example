// https://github.com/usbarmory/tamago-example
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package network

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"html"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/usbarmory/tamago-example/cmd"
)

func generateTLSCerts(address net.IP) ([]byte, []byte, error) {
	TLSCert := new(bytes.Buffer)
	TLSKey := new(bytes.Buffer)

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<63-1))

	log.Printf("generating TLS keypair IP: %s, Serial: %X", address.String(), serial)

	validFrom, _ := time.Parse(time.RFC3339, "1981-01-07T00:00:00Z")
	validUntil, _ := time.Parse(time.RFC3339, "2022-01-07T00:00:00Z")

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{"TamaGo Example"},
			OrganizationalUnit: []string{"TamaGo test certificates"},
			CommonName:         address.String(),
		},
		IPAddresses:        []net.IP{address},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		PublicKeyAlgorithm: x509.ECDSA,
		NotBefore:          validFrom,
		NotAfter:           validUntil,
		SubjectKeyId:       []byte{1, 2, 3, 4, 5},
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	caTemplate := certTemplate
	caTemplate.SerialNumber = serial
	caTemplate.SubjectKeyId = []byte{1, 2, 3, 4, 6}
	caTemplate.BasicConstraintsValid = true
	caTemplate.IsCA = true
	caTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	caTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub := &priv.PublicKey
	cert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &caTemplate, pub, priv)

	if err != nil {
		return nil, nil, err
	}

	pem.Encode(TLSCert, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	ecb, _ := x509.MarshalECPrivateKey(priv)
	pem.Encode(TLSKey, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecb})

	h := sha256.New()
	h.Write(cert)

	log.Printf("SHA-256 fingerprint: % X", h.Sum(nil))

	return TLSCert.Bytes(), TLSKey.Bytes(), nil
}

func flushingHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "Fri, 07 Jan 1981 00:00:00 GMT")

		journal.Sync()

		h.ServeHTTP(w, r)
	}
}

func setupStaticWebAssets() {
	file, err := os.OpenFile("/index.html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("<html><body>")
	file.WriteString(fmt.Sprintf("<p>%s</p><ul>", html.EscapeString(cmd.Banner)))
	file.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, "/tamago-example.log", "/tamago-example.log"))
	file.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, "/dir", "/dir"))
	file.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, "/debug/pprof", "/debug/pprof"))
	file.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, "/debug/statsviz", "/debug/statsviz"))
	file.WriteString("</ul></body></html>")

	static := http.FileServer(http.Dir("/"))
	staticHandler := flushingHandler(static)
	http.Handle("/", http.StripPrefix("/", staticHandler))
}

func startWebServer(listener net.Listener, addr string, port uint16, https bool) {
	var err error

	srv := &http.Server{
		Addr: addr + ":" + fmt.Sprintf("%d", port),
	}

	if https {
		TLSCert, TLSKey, err := generateTLSCerts(net.ParseIP(addr))

		if err != nil {
			log.Fatal("TLS cert|key error: ", err)
		}

		log.Printf("generated TLS certificate:\n%s", TLSCert)
		log.Printf("generated TLS key:\n%s", TLSKey)

		certificate, err := tls.X509KeyPair(TLSCert, TLSKey)

		if err != nil {
			log.Fatal("X509KeyPair error: ", err)
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
	}

	log.Printf("starting web server at %s:%d", addr, port)

	if https {
		err = srv.ServeTLS(listener, "", "")
	} else {
		err = srv.Serve(listener)
	}

	log.Fatal("server returned unexpectedly ", err)
}
