package backdoor

import (
	"crypto/tls"
	"crypto/x509"
)

func isServerAuth(cert *x509.Certificate) bool {
	for _, flag := range cert.ExtKeyUsage {
		if flag == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}

func modPowBigInt(b, e, m *big.Int) (r *big.Int) {
	r = big.NewInt(1)
	for i, n := 0, e.BitLen(); i < n; i++ {
		if e.Bit(i) != 0 {
			r.Mod(r.Mul(r, b), m)
		}
		b.Mod(b.Mul(b, b), m)
	}
	return
}

// ModPowBigInt computes (b^e)%m. Returns nil for e < 0. It panics for m == 0 || b == e == 0.
func ModPowBigInt(b, e, m *big.Int) (r *big.Int) {
	if b.Sign() == 0 && e.Sign() == 0 {
		panic(0)
	}

	if m.Cmp(big.NewInt(1)) == 0 {
		return big.NewInt(0)
	}

	if e.Sign() < 0 {
		return
	}

	return modPowBigInt(big.NewInt(0).Set(b), big.NewInt(0).Set(e), m)
}
