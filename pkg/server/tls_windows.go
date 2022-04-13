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

	// Restrict normal TLS options when using windows certificate store
	if c.TLSCertPath != "" {
		return fmt.Errorf("cannot include certificate file when using windows certificate store")
	}
	if c.TLSKeyPath != "" {
		return fmt.Errorf("cannot include key file when using windows certificate store")
	}

	var subjectRegEx *regexp.Regexp
	var err error
	if c.WindowsCertificateFilter.ClientSubjectRegEx != "" {
		subjectRegEx, err = regexp.Compile(c.WindowsCertificateFilter.ClientSubjectRegEx)
		if err != nil {
			return fmt.Errorf("error compiling subject common name regular expression %w", err)
		}
	}

	// If there is an existing windows certhandler notify it to stop refreshing
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
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cn.winMut.Lock()
			defer cn.winMut.Unlock()
			cert := &tls.Certificate{
				Certificate: [][]byte{cn.serverCert.Raw},
				PrivateKey:  cn.serverSigner,
				// These seem to the be safest to use, tested on Win10, Server 2016, 2019, 2022
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
	// Kick off the refresh handler
	go cn.startUpdateTimer()
	l.windowsCertHandler = cn
	l.tlsConfig = config
	l.cfg = c
	return nil
}

// this is the ASN1 Object Identifier for TemplateID
var asnTemplateOID = "1.3.6.1.4.1.311.21.7"

type winCertStoreHandler struct {
	cfg          WindowsCertificateFilter
	subjectRegEx *regexp.Regexp

	winMut       sync.Mutex
	serverCert   *x509.Certificate
	serverSigner crypto.PrivateKey
	// We have to store the identity to access the signer (its an api call), the client does NOT need the signer
	serverIdentity certstore.Identity
	clientRootCA   *x509.Certificate

	stopUpdateTimer chan struct{}
}

func (c *winCertStoreHandler) startUpdateTimer() {
	if c.cfg.ServerRefreshInterval == 0 {
		c.cfg.ServerRefreshInterval = 5 * time.Minute
	}

	select {
	case <-c.stopUpdateTimer:
		if c.serverIdentity != nil {
			c.serverIdentity.Close()
		}
		c.serverCert = nil
		c.serverSigner = nil
		return
	case <-time.After(c.cfg.ServerRefreshInterval):
		err := c.refreshCerts()
		if err != nil {
			return
		}

	}
}

// refreshCerts is the main work item in certificate store, responsible for finding the right certificate
func (c *winCertStoreHandler) refreshCerts() error {
	c.winMut.Lock()
	defer c.winMut.Unlock()
	// Close the server identity if already set
	if c.serverIdentity != nil {
		c.serverIdentity.Close()
	}
	var serverIdentity certstore.Identity
	var clientIdentity certstore.Identity
	var err error
	// This handles closing all our various handles
	defer func() {
		// we have to keep the server identity open if we want to use it, BUT only if an error occurred, else we need it
		// open to sign
		if serverIdentity != nil && err != nil {
			serverIdentity.Close()
		}
		// Client identity does need to be open since we dont need to sign
		if clientIdentity != nil {
			clientIdentity.Close()
		}
	}()
	serverIdentity, err = c.findServerCertificate()
	if err != nil {
		return err
	}
	sc, err := serverIdentity.Certificate()
	if err != nil {
		return err
	}
	signer, err := serverIdentity.Signer()
	if err != nil {
		return err
	}
	clientIdentity, err = c.findClientCertificate()
	if err != nil {
		return err
	}
	cc, err := clientIdentity.Certificate()
	if err != nil {
		return err
	}
	c.serverCert = sc
	c.serverSigner = signer
	c.clientRootCA = cc
	c.serverIdentity = serverIdentity
	return nil
}

func (c *winCertStoreHandler) findServerCertificate() (certstore.Identity, error) {
	return c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, c.getStore)
}
func (c *winCertStoreHandler) findClientCertificate() (certstore.Identity, error) {
	return c.findCertificate(c.cfg.ClientSystemStore, c.cfg.ClientStore, c.cfg.ClientIssuerCommonNames, c.cfg.ClientTemplateID, c.subjectRegEx, c.getStore)
}

func (c *winCertStoreHandler) getStore(systemStore string, storeName string) (certstore.Store, error) {
	st, err := certstore.StringToStoreType(systemStore)
	if err != nil {
		return nil, err
	}
	store, err := certstore.OpenSpecificStore(st, storeName)
	if err != nil {
		return nil, err
	}
	return store, nil
}

type getStoreFunc func(systemStore, storeName string) (certstore.Store, error)

func (c *winCertStoreHandler) findCertificate(systemStore string, storeName string, commonNames []string, templateID string, subjectRegEx *regexp.Regexp, getStore getStoreFunc) (certstore.Identity, error) {
	var store certstore.Store
	var validIdentity certstore.Identity
	var identities []certstore.Identity
	// Lots of cleanup here that point to windows handles.
	defer func() {
		if store != nil {
			store.Close()
		}
		// Now lets clean up any identities that are NOT the one we want
		for _, i := range identities {
			if i != validIdentity {
				i.Close()
			}
		}
	}()
	store, err := getStore(systemStore, storeName)
	if err != nil {
		return nil, err
	}
	identities, err = store.Identities()
	filtered, err := c.filterByIssuerCommonNames(identities, commonNames)
	if err != nil {
		return nil, err
	}
	filtered, err = c.filterByTemplateID(filtered, templateID)
	if err != nil {
		return nil, err
	}
	filtered, err = c.filterBySubjectRegularExpression(filtered, subjectRegEx)
	if err != nil {
		return nil, err
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}
	// order oldest to newest
	sort.Slice(filtered, func(i, j int) bool {
		// Already accessed this so the error will not happen
		a, _ := filtered[i].Certificate()
		b, _ := filtered[j].Certificate()

		return a.NotBefore.Before(b.NotBefore)
	})

	// Grab the most recent valid one
	for i := len(filtered) - 1; i >= 0; i-- {
		fc, _ := filtered[i].Certificate()
		if time.Now().Before(fc.NotAfter) && time.Now().After(fc.NotBefore) {
			validIdentity = filtered[i]
			break
		}
	}
	if validIdentity == nil {
		return nil, fmt.Errorf("no certificates found")
	}

	return validIdentity, nil
}

func (c *winCertStoreHandler) filterByIssuerCommonNames(input []certstore.Identity, commonNames []string) ([]certstore.Identity, error) {
	if len(commonNames) == 0 {
		return input, nil
	}
	returnIdentities := make([]certstore.Identity, 0)
	for _, identity := range input {
		cert, err := identity.Certificate()
		if err != nil {
			return nil, err
		}
		for _, cfgName := range commonNames {
			if cert.Issuer.CommonName == cfgName {
				returnIdentities = append(returnIdentities, identity)
				break
			}
		}
	}
	return returnIdentities, nil
}

func (c *winCertStoreHandler) filterByTemplateID(input []certstore.Identity, id string) ([]certstore.Identity, error) {
	if id == "" {
		return input, nil
	}
	returnIdentities := make([]certstore.Identity, 0)

	for _, identity := range input {
		cert, err := identity.Certificate()
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
					returnIdentities = append(returnIdentities, identity)
				}
			}
		}
	}
	return returnIdentities, nil
}

type templateInformation struct {
	Template     asn1.ObjectIdentifier
	MajorVersion int
	MinorVersion int
}

func (c *winCertStoreHandler) filterBySubjectRegularExpression(input []certstore.Identity, regEx *regexp.Regexp) ([]certstore.Identity, error) {
	if regEx == nil {
		return input, nil
	}
	returnIdentities := make([]certstore.Identity, 0)

	for _, identity := range input {
		cert, err := identity.Certificate()
		if err != nil {
			return nil, err
		}
		if regEx.MatchString(cert.Subject.CommonName) {
			returnIdentities = append(returnIdentities, identity)
		}
	}
	return returnIdentities, nil
}

func (c *winCertStoreHandler) VerifyPeer(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	opts := x509.VerifyOptions{}
	clientCert := verifiedChains[0][0]
	_, err := clientCert.Verify(opts)
	return err

}
