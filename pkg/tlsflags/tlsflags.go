package tlsflags

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"k8s.io/klog"
)

func Bind(flags FlagSet) (f *Flags) {
	f = &Flags{}
	f.Bind(flags)
	return
}

type Flags struct {
	KeyFile,
	CertFile,
	CAFile string
}

// FlagSet matches flag.FlagSet and pflag.FlagSet
type FlagSet interface {
	DurationVar(varPtr *time.Duration, name string, value time.Duration, doc string)
	IntVar(varPtr *int, name string, value int, doc string)
	StringVar(varPtr *string, name, value, doc string)
	Uint64Var(varPtr *uint64, name string, value uint64, doc string)
}

func (f *Flags) Bind(flags FlagSet) {
	flags.StringVar(&f.KeyFile, "tls-key", "", "TLS key file")
	flags.StringVar(&f.CertFile, "tls-crt", "", "TLS certificate file")
	flags.StringVar(&f.CAFile, "tls-ca", "", "TLS CA certificate file")
}

func (f *Flags) Config() (cfg *tls.Config) {
	if f == nil || f.CAFile == "" && f.KeyFile == "" && f.CertFile == "" {
		return
	}

	cfg = &tls.Config{}

	if f.KeyFile != "" || f.CertFile != "" {
		cert, err := tls.LoadX509KeyPair(f.CertFile, f.KeyFile)
		if err != nil {
			klog.Fatal("failed to load TLS key pair: ", err)
		}

		cfg.Certificates = []tls.Certificate{cert}
	}

	if f.CAFile != "" {
		data, err := ioutil.ReadFile(f.CAFile)
		if err != nil {
			klog.Fatal("failed to load TLS CA certificate: ", err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(data) {
			klog.Fatal("failed to parse CA certificate")
		}

		cfg.RootCAs = pool
	}

	return
}
