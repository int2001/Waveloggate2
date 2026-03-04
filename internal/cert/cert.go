package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Paths holds the filesystem paths to the certificate and key files.
type Paths struct {
	Key  string
	Cert string
}

// Info describes the current certificate state.
type Info struct {
	CertPath    string `json:"certPath"`
	Exists      bool   `json:"exists"`
	IsInstalled bool   `json:"isInstalled"`
}

// InstallResult is the outcome of a certificate installation attempt.
type InstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Command string `json:"command"` // manual fallback command for the user
}

// certDir returns (and creates) the directory where certs are stored.
func certDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "WavelogGate", "certs")
	return dir, os.MkdirAll(dir, 0o700)
}

// Setup loads an existing certificate/key pair or generates a new one.
// Regenerates if the files are missing, unparseable, mismatched, or expired.
// Returns (paths, newlyGenerated, error).
func Setup() (Paths, bool, error) {
	dir, err := certDir()
	if err != nil {
		return Paths{}, false, err
	}

	p := Paths{
		Key:  filepath.Join(dir, "server.key"),
		Cert: filepath.Join(dir, "server.crt"),
	}

	if certValid(p) {
		return p, false, nil
	}

	if err := generate(p); err != nil {
		return Paths{}, false, err
	}
	return p, true, nil
}

// certValid returns true when the cert and key files exist, can be loaded as a
// valid TLS pair, and the certificate has not yet expired.
func certValid(p Paths) bool {
	tlsCert, err := tls.LoadX509KeyPair(p.Cert, p.Key)
	if err != nil {
		return false
	}
	if len(tlsCert.Certificate) == 0 {
		return false
	}
	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return false
	}
	return time.Now().Before(x509Cert.NotAfter)
}

// generate creates a self-signed ECDSA P-256 certificate valid for 10 years.
func generate(p Paths) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "127.0.0.1",
			Organization: []string{"WavelogGate"},
		},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Write key.
	keyFile, err := os.OpenFile(p.Key, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return err
	}

	// Write cert.
	certFile, err := os.OpenFile(p.Cert, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer certFile.Close()
	return pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

// IsCertInstalled reports whether the certificate is trusted by the OS.
func IsCertInstalled(certPath string) bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("security", "find-certificate", "-c", "127.0.0.1",
			"-p", "/Library/Keychains/System.keychain").Output()
		return err == nil && len(out) > 0
	case "windows":
		out, err := exec.Command("certutil", "-store", "Root").Output()
		return err == nil && strings.Contains(string(out), "127.0.0.1")
	default: // Linux
		anchors := []string{
			"/usr/local/share/ca-certificates/",
			"/etc/pki/ca-trust/source/anchors/",
			"/etc/ca-certificates/trust-source/anchors/",
		}
		for _, dir := range anchors {
			if _, err := os.Stat(filepath.Join(dir, "waveloggate.crt")); err == nil {
				return true
			}
		}
		return false
	}
}

// Install attempts to add the certificate to the system trust store.
func Install(certPath string) InstallResult {
	switch runtime.GOOS {
	case "darwin":
		script := `do shell script "security add-trusted-cert -d -p ssl -p basic -k /Library/Keychains/System.keychain '` +
			certPath + `'" with administrator privileges`
		cmd := exec.Command("osascript", "-e", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return InstallResult{
				Success: false,
				Message: "Installation failed: " + strings.TrimSpace(string(out)),
				Command: "sudo security add-trusted-cert -d -p ssl -p basic -k /Library/Keychains/System.keychain \"" + certPath + "\"",
			}
		}
		return InstallResult{Success: true, Message: "Certificate installed. Please restart your browser."}

	case "windows":
		cmd := exec.Command("certutil", "-addstore", "-f", "Root", certPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			psCmd := `Start-Process certutil -ArgumentList '-addstore','-f','Root','` + certPath + `' -Verb RunAs`
			return InstallResult{
				Success: false,
				Message: "Auto-install failed. Run the command below as Administrator:",
				Command: "powershell -Command \"" + psCmd + "\"",
			}
		}
		_ = out
		return InstallResult{Success: true, Message: "Certificate installed. Please restart your browser."}

	default: // Linux — manual only
		return InstallResult{
			Success: false,
			Message: "Automatic installation is not supported on Linux. Run the command below:",
			Command: "sudo cp \"" + certPath + "\" /usr/local/share/ca-certificates/waveloggate.crt && sudo update-ca-certificates",
		}
	}
}

// GetInfo returns the current certificate information.
func GetInfo(certPath string) Info {
	_, err := os.Stat(certPath)
	exists := err == nil
	return Info{
		CertPath:    certPath,
		Exists:      exists,
		IsInstalled: exists && IsCertInstalled(certPath),
	}
}
