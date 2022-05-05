package server

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/github/smimesign/certstore"
)

func (l *tlsListener) applyWindowsCertificateStore(c TLSConfig) error {

	// Restrict normal TLS options when using windows certificate store
	if c.TLSCertPath != "" {
		return fmt.Errorf("at most one of cert_file and windows_certificate_filter can be configured")
	}
	if c.TLSKeyPath != "" {
		return fmt.Errorf("at most one of cert_key and windows_certificate_filter can be configured")
	}
	if c.WindowsCertificateFilter.Server == nil {
		return fmt.Errorf("windows certificate filter requires a server block defined")
	}

	var subjectRegEx *regexp.Regexp
	var err error
	if c.WindowsCertificateFilter.Client != nil && c.WindowsCertificateFilter.Client.SubjectRegEx != "" {
		subjectRegEx, err = regexp.Compile(c.WindowsCertificateFilter.Client.SubjectRegEx)
		if err != nil {
			return fmt.Errorf("error compiling subject common name regular expression: %w", err)
		}
	}

	// If there is an existing windows certhandler notify it to stop refreshing
	if l.cancelWindowsCert != nil {
		l.cancelWindowsCert()
		l.cancelWindowsCert = nil
	}
	cancelCtx := context.Background()
	cancelCtx, cancelHandler := context.WithCancel(cancelCtx)
	l.cancelWindowsCert = cancelHandler
	cn := &winCertStoreHandler{
		cfg:           *c.WindowsCertificateFilter,
		subjectRegEx:  subjectRegEx,
		cancelContext: cancelCtx,
		log:           l.log,
	}
	err = cn.refreshCerts()
	if err != nil {
		return err
	}
	// CertPool, if there is a clientrootca use it, else pass in nil which will use the default certs
	var certPool *x509.CertPool
	if cn.clientRootCA != nil {
		certPool = x509.NewCertPool()
		certPool.AddCert(cn.clientRootCA)
	}

	config := &tls.Config{
		ClientCAs:             certPool,
		VerifyPeerCertificate: cn.VerifyPeer,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cn.winMut.Lock()
			defer cn.winMut.Unlock()
			cert := &tls.Certificate{
				Certificate: [][]byte{cn.serverCert.Raw},
				PrivateKey:  cn.serverSigner,
				Leaf:        cn.serverCert,
				// These seem to the be safest to use, tested on Win10, Server 2016, 2019, 2022
				SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.PKCS1WithSHA512,
					tls.PKCS1WithSHA384,
					tls.PKCS1WithSHA256,
				},
			}
			return cert, nil
		},
		// Windows has broad support for 1.2, only 2022 has support for 1.3
		MaxVersion: tls.VersionTLS12,
	}

	ca, err := getClientAuthFromString(c.ClientAuth)
	if err != nil {
		return err
	}
	config.ClientAuth = ca
	cn.clientAuth = ca
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
	log          log.Logger

	winMut       sync.Mutex
	serverCert   *x509.Certificate
	serverSigner crypto.PrivateKey
	// We have to store the identity to access the signer (it's a win32 api call), if we close the identity
	// we lose access to the signer
	// the client does NOT need the signer
	serverIdentity certstore.Identity
	clientRootCA   *x509.Certificate
	clientAuth     tls.ClientAuthType

	cancelContext context.Context
}

func (c *winCertStoreHandler) startUpdateTimer() {
	refreshInterval := 5 * time.Minute
	c.winMut.Lock()
	if c.cfg.Server.RefreshInterval != 0 {
		refreshInterval = c.cfg.Server.RefreshInterval
	}
	c.winMut.Unlock()
	for {
		select {
		case <-c.cancelContext.Done():
			if c.serverIdentity != nil {
				c.serverIdentity.Close()
			}
			c.serverCert = nil
			c.serverSigner = nil
			return
		case <-time.After(refreshInterval):
			err := c.refreshCerts()
			if err != nil {
				level.Error(c.log).Log("msg", "error refreshing Windows certificates", "err", err)
			}

		}
	}
}

// refreshCerts is the main work item in certificate store, responsible for finding the right certificate
func (c *winCertStoreHandler) refreshCerts() (err error) {
	c.winMut.Lock()
	defer c.winMut.Unlock()
	level.Debug(c.log).Log("msg", "refreshing Windows certificates")
	// Close the server identity if already set
	if c.serverIdentity != nil {
		c.serverIdentity.Close()
	}
	var serverIdentity certstore.Identity
	var clientIdentity certstore.Identity
	var clientCertificate *x509.Certificate
	// This handles closing all our various handles
	defer func() {
		// we have to keep the server identity open if we want to use it, BUT only if an error occurred, else we need it
		// open to sign
		if serverIdentity != nil && err != nil {
			serverIdentity.Close()
		}
		// Client identity does NOT need to be open since we dont need to sign
		if clientIdentity != nil {
			clientIdentity.Close()
		}
	}()
	serverIdentity, err = c.findServerIdentity()
	if err != nil {
		return fmt.Errorf("failed finding server identity %w", err)
	}
	sc, err := serverIdentity.Certificate()
	if err != nil {
		return fmt.Errorf("failed getting server certificate %w", err)
	}
	signer, err := serverIdentity.Signer()
	if err != nil {
		return fmt.Errorf("failed getting server signer %w", err)

	}
	clientIdentity, err = c.findClientIdentity()
	if err != nil {
		return fmt.Errorf("failed getting client identity %w", err)
	}
	if clientIdentity == nil && (c.clientAuth == tls.RequireAndVerifyClientCert || c.clientAuth == tls.RequestClientCert) {
		return fmt.Errorf("client auth requires a certificate (RequireAndVerifyClientCert or RequestClientCert) and failed getting client identity")
	}
	if clientIdentity != nil {
		clientCertificate, err = clientIdentity.Certificate()
		if err != nil {
			return fmt.Errorf("failed getting client certificate %w", err)
		}
	}

	c.serverCert = sc
	c.serverSigner = signer
	c.clientRootCA = clientCertificate
	c.serverIdentity = serverIdentity
	return
}

func (c *winCertStoreHandler) findServerIdentity() (certstore.Identity, error) {
	return c.findCertificate(c.cfg.Server.SystemStore, c.cfg.Server.Store, c.cfg.Server.IssuerCommonNames, c.cfg.Server.TemplateID, nil, c.getStore)
}
func (c *winCertStoreHandler) findClientIdentity() (certstore.Identity, error) {
	// Client certificate is not required
	// Valid configurations are checked further in findCertificate
	if c.cfg.Client == nil {
		return nil, nil
	}
	return c.findCertificate(c.cfg.Client.SystemStore, c.cfg.Client.Store, c.cfg.Client.IssuerCommonNames, c.cfg.Client.TemplateID, c.subjectRegEx, c.getStore)
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
	sort.Slice(filtered, func(certI, certJ int) bool {
		// Already accessed this so the error will not happen
		a, _ := filtered[certI].Certificate()
		b, _ := filtered[certJ].Certificate()

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
