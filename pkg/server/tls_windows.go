package server

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/github/smimesign/certstore"
)

func (l *tlsListener) applyWindowsCertificateStore(c TLSConfig) error {

	if c.TLSCertPath != "" {
		return fmt.Errorf("cannot include certificate file when using windows certificate store")
	}
	if c.TLSKeyPath != "" {
		return fmt.Errorf("cannot include key file when using windows certificate store")
	}

	var subjectRegEx *regexp.Regexp
	var err error
	if c.WindowsCertificateFilter.ClientSubjectCommonName != "" {
		subjectRegEx, err = regexp.Compile(c.WindowsCertificateFilter.ClientSubjectCommonName)
		if err != nil {
			return fmt.Errorf("error compiling subject common name regular expression %w", err)
		}
	}
	if l.windowsCertHandler != nil {
		l.windowsCertHandler.stopUpdateTimer <- struct{}{}
	}

	cn := &winCertStoreHandler{
		cfg:             *c.WindowsCertificateFilter,
		subjectRegEx:    subjectRegEx,
		stopUpdateTimer: make(chan struct{}, 1),
	}
	err = cn.refreshCerts()
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(cn.clientRootCA)
	config := &tls.Config{
		VerifyPeerCertificate: cn.VerifyPeer,
		ClientAuth:            tls.RequireAndVerifyClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cn.winMut.Lock()
			defer cn.winMut.Unlock()

			cert := &tls.Certificate{
				Certificate: [][]byte{cn.serverCert.Raw},
				PrivateKey:  cn.serverSigner,
				SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.PKCS1WithSHA256,
					tls.PKCS1WithSHA384,
					tls.PKCS1WithSHA512,
				},
			}

			return cert, nil
		},
		// Windows has broad support for 1.2, only 2022 has support for 1.3
		MaxVersion: tls.VersionTLS12,
	}

	switch c.ClientAuth {
	case "RequestClientCert":
		config.ClientAuth = tls.RequestClientCert
	case "RequireAnyClientCert", "RequireClientCert": // Preserved for backwards compatibility.
		config.ClientAuth = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		config.ClientAuth = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		config.ClientAuth = tls.RequireAndVerifyClientCert
	case "", "NoClientCert":
		config.ClientAuth = tls.NoClientCert
	default:
		return fmt.Errorf("invalid ClientAuth %q", c.ClientAuth)
	}
	go cn.startUpdateTimer()
	l.windowsCertHandler = cn
	l.tlsConfig = config
	l.cfg = c
	return nil
}

var asnTemplateOID = "1.3.6.1.4.1.311.21.7"

type winCertStoreHandler struct {
	cfg          WindowsCertificateFilter
	subjectRegEx *regexp.Regexp

	winMut         sync.Mutex
	serverIdentity certstore.Identity
	serverCert     *x509.Certificate
	serverSigner   crypto.PrivateKey
	serverStore    certstore.Store
	clientRootCA   *x509.Certificate

	stopUpdateTimer chan struct{}
}

func (c *winCertStoreHandler) closeHandles() {
	if c.serverIdentity != nil {
		c.serverIdentity.Close()
		c.serverIdentity = nil
	}
	if c.serverStore != nil {
		c.serverStore.Close()
		c.serverStore = nil
	}
	c.serverSigner = nil
	c.serverCert = nil
}

func (c *winCertStoreHandler) startUpdateTimer() {
	if c.cfg.ServerRefreshInterval == 0 {
		c.cfg.ServerRefreshInterval = 5 * time.Minute
	}

	select {
	case <-c.stopUpdateTimer:
		c.closeHandles()
		return
	case <-time.After(c.cfg.ServerRefreshInterval):
		err := c.refreshCerts()
		if err != nil {
			return
		}

	}
}

func (c *winCertStoreHandler) refreshCerts() error {
	c.winMut.Lock()
	defer c.winMut.Unlock()
	serverIdentity, serverStore, err := c.findServerCertificate()
	if err != nil {
		c.closeHandles()
		return err
	}
	sc, err := serverIdentity.Certificate()
	if err != nil {
		c.closeHandles()
		return err
	}
	signer, err := serverIdentity.Signer()
	if err != nil {
		c.closeHandles()
		return err
	}
	clientIdentity, clientStore, err := c.findClientCertificate()
	if err != nil {
		clientIdentity.Close()
		c.closeHandles()
		return err
	}
	cc, err := clientIdentity.Certificate()
	if err != nil {
		clientStore.Close()
		clientIdentity.Close()
		c.closeHandles()
		return err
	}
	// Close any existing handles before we assign new handles
	c.closeHandles()
	c.serverCert = sc
	c.serverSigner = signer
	c.serverStore = serverStore
	c.serverIdentity = serverIdentity
	c.clientRootCA = cc
	clientStore.Close()
	clientIdentity.Close()
	return nil
}

func (c *winCertStoreHandler) findServerCertificate() (certstore.Identity, certstore.Store, error) {
	return c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID)
}
func (c *winCertStoreHandler) findClientCertificate() (certstore.Identity, certstore.Store, error) {
	return c.findCertificate(c.cfg.ClientSystemStore, c.cfg.ClientStore, c.cfg.ClientIssuerCommonNames, c.cfg.ClientTemplateID)
}

func (c *winCertStoreHandler) findCertificate(systemStore string, storeName string, commonNames []string, templateID string) (certstore.Identity, certstore.Store, error) {
	st, err := certstore.StringToStoreType(systemStore)
	if err != nil {
		return nil, nil, err
	}
	store, err := certstore.OpenSpecificStore(st, storeName)
	if err != nil {
		return nil, nil, err
	}
	identities, err := store.Identities()
	filtered, err := c.filterByIssuerName(identities, commonNames)
	if err != nil {
		return nil, nil, err
	}
	filtered, err = c.filterByTemplateID(filtered, templateID)
	if err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	if len(filtered) == 0 {
		return nil, nil, fmt.Errorf("no certificates found")
	}
	sort.Slice(filtered, func(i, j int) bool {
		// Already accessed this so the error will not happen
		c, _ := filtered[i].Certificate()
		b, _ := filtered[j].Certificate()

		return c.NotBefore.Before(b.NotBefore)
	})

	var validStore certstore.Identity
	for i := 0; i < len(filtered); {
		c, _ := filtered[i].Certificate()
		if time.Now().Before(c.NotAfter) && time.Now().After(c.NotBefore) {
			validStore = filtered[i]
			break
		}
	}
	// Now lets clean up any identities we did NOT use
	for _, i := range identities {
		if i != validStore {
			i.Close()
		}
	}
	if validStore == nil {
		return nil, nil, fmt.Errorf("no certificates found")
	}
	return validStore, store, nil
}

func (c *winCertStoreHandler) filterByIssuerName(input []certstore.Identity, commonNames []string) ([]certstore.Identity, error) {
	if len(commonNames) == 0 {
		return input, nil
	}
	returnCerts := make([]certstore.Identity, 0)
	for _, c := range input {
		cert, err := c.Certificate()
		if err != nil {
			return nil, err
		}
		for _, cfgName := range commonNames {
			if cert.Issuer.CommonName == cfgName {
				returnCerts = append(returnCerts, c)
				break
			}
		}
	}
	return returnCerts, nil
}

func (c *winCertStoreHandler) filterByTemplateID(input []certstore.Identity, id string) ([]certstore.Identity, error) {
	if id == "" {
		return input, nil
	}
	returnCerts := make([]certstore.Identity, 0)

	for _, c := range input {
		cert, err := c.Certificate()
		if err != nil {
			return nil, err
		}
		for _, ext := range cert.Extensions {
			if ext.Id.String() == asnTemplateOID {
				templateInfo := &templateInformation{}
				_, err := asn1.Unmarshal(ext.Value, templateInfo)
				if err != nil {
					return nil, err
				}
				if templateInfo.Template.String() == id {
					returnCerts = append(returnCerts, c)
				}
			}
		}
	}
	return returnCerts, nil
}

func (c *winCertStoreHandler) VerifyPeer(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	opts := x509.VerifyOptions{}
	clientCert := verifiedChains[0][0]
	_, err := clientCert.Verify(opts)
	return err

}

type templateInformation struct {
	Template     asn1.ObjectIdentifier
	MajorVersion int
	MinorVersion int
}
