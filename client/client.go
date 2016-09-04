package main

import (
	//	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	//	"io"
	"io/ioutil"
	//	"os"
)

const (
	CA_crt = "./ca.crt"
	BM_crt = "./botmaster.crt"
)

func printTLSStatus(c *tls.Conn) {
	status := c.ConnectionState()
	fmt.Printf("===== TLS Conn Status =====\n")
	fmt.Printf("Server Addr: %s\n", c.RemoteAddr())
	fmt.Printf("Local Addr: %s\n", c.LocalAddr())
	fmt.Printf("TLS Version: %X\n", status.Version)
	fmt.Printf("Cipher: %X\n", status.CipherSuite)
	fmt.Println("HandshakeComplete:", status.HandshakeComplete)
	fmt.Println("TLSUnique:", status.TLSUnique)

	block := pem.Block{
		Bytes: status.PeerCertificates[0].Raw,
		Type:  "CERTIFICATE",
	}
	fmt.Printf("PeerCertificates:\n %s", pem.EncodeToMemory(&block))

	fmt.Printf("===========================\n")
}

func getClientConfig() *tls.Config {
	PEMCAcert, err := ioutil.ReadFile(CA_crt)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(PEMCAcert) {
		panic("Failed appending CA cert")
	}

	config := &tls.Config{
		RootCAs: certPool,

		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
		PreferServerCipherSuites: true,

		InsecureSkipVerify:     false,
		SessionTicketsDisabled: true,
		//Use only TLS v1.2
		MinVersion: tls.VersionTLS12,
	}
	return config
}

func isServerAuth(cert *x509.Certificate) bool { //FIXME: Use libbackdoor
	for _, flag := range cert.ExtKeyUsage {
		if flag == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}

//func PEM2PubKey(pathname string) rsa.PublicKey {
//	PEMcert, err := ioutil.ReadFile(pathname)
//	if err != nil {
//		panic(err)
//	}
//	certificate, _ := x509.ParseCertificate(PEMcert)
//	return certificate.PublicKey
//	//pubkey, _ := x509.MarshalPKIXPublicKey(certificate.PublicKey)
//	//return pubkey
//}

func main() {
	config := getClientConfig()

	// Load BotMaster certificate
	//	PEMBMcert, err := ioutil.ReadFile(BM_crt)
	//	if err != nil {
	//		panic(err)
	//	}
	//	BMcertificate, _ := x509.ParseCertificate(PEMBMcert)
	//	BMpub := BMcertificate.PublicKey

	conn, err := tls.Dial("tcp", "localhost:51000", config)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if !isServerAuth(conn.ConnectionState().PeerCertificates[0]) {
		//FIXME: Do not connect to server if go here, it could be an ephemeral server
		fmt.Printf("...::: Certificate not valid for a C2s :::...\n\n")
	}

	err = conn.Handshake()
	if err != nil {
		fmt.Printf("Failed handshake:%v\n", err)
		return
	}

	//_, err = io.Copy(buff, conn)
	result, err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Printf("Failed receiving data:%v\n", err)
	}

	fmt.Printf("%s", result)
	//printTLSStatus(conn)
}
