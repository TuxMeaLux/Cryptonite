package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

const (
	certificateFile = "./server.crt"
	privateKeyFile  = "./server.key"
	ca_crt          = "./ca.crt"
)

func isServerAuth(cert *x509.Certificate) bool { //FIXME: Use libbackdoor
	for _, flag := range cert.ExtKeyUsage {
		if flag == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}

func getServerConfig() *tls.Config {
	serverCert, err := tls.LoadX509KeyPair(certificateFile, privateKeyFile)
	if err != nil {
		panic(err)
	}
	serverCert.Leaf, _ = x509.ParseCertificate(serverCert.Certificate[0])
	if !isServerAuth(serverCert.Leaf) {
		//FIXME: Do not start server if go here
		fmt.Printf("...::: Certificate not valid for a C2s :::...\n\n")
	}
	serverCert_array := make([]tls.Certificate, 1)
	serverCert_array[0] = serverCert

	PEMCAcert, err := ioutil.ReadFile(ca_crt)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(PEMCAcert) {
		panic("Failed appending CA cert")
	}

	config := &tls.Config{
		Certificates:      serverCert_array,
		RootCAs:           certPool,
		NameToCertificate: nil,

		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},

		PreferServerCipherSuites: true,

		//Use only TLS v1.2
		MinVersion: tls.VersionTLS12,

		//Don't allow session resumption
		SessionTicketsDisabled: true,
	}
	return config
}

func main() {
	config := getServerConfig()

	listener, err := tls.Listen("tcp", "localhost:51000", config)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Printf("[+] Server started\n")
	for {
		fmt.Printf("[+] Listening...\n")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			fmt.Printf("[+] Connection Accepted form %s\n", c.RemoteAddr())
			fmt.Fprintf(c, "*************************\n")
			fmt.Fprintf(c, "* SURPRISE MOTHERFUCKER *\n")
			fmt.Fprintf(c, "*************************\n\n")
			if err != nil {
				fmt.Printf("Error on connection:%v\n", err)
			}
			c.Close()
		}(conn)
	}
}
