package server

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"github.com/github/smimesign/certstore"
	"github.com/stretchr/testify/require"
	"math/big"
	"regexp"
	"testing"
	"time"
)

func TestEasyFilter(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:       "My",
			ServerSystemStore: "LocalMachine",
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	serverIdentity, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)
	require.NoError(t, err)
	foundCert, err := serverIdentity.Certificate()
	require.NoError(t, err)
	require.NotNil(t, serverIdentity)
	require.True(t, foundCert == sc)
}

func TestTemplateIDFilter(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:       "My",
			ServerSystemStore: "LocalMachine",
			ServerTemplateID:  "1.2.3",
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	serverIdentity, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)
	require.NoError(t, err)
	foundCert, err := serverIdentity.Certificate()
	require.NoError(t, err)
	require.NotNil(t, serverIdentity)
	require.True(t, foundCert == sc)
}

func TestCommonName(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:             "My",
			ServerSystemStore:       "LocalMachine",
			ServerTemplateID:        "1.2.3",
			ServerIssuerCommonNames: []string{"TEST"},
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "TEST", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	serverIdentity, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)
	require.NoError(t, err)
	foundCert, err := serverIdentity.Certificate()
	require.NoError(t, err)
	require.NotNil(t, serverIdentity)
	require.True(t, foundCert == sc)
}

func TestCommonName_Fail(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:             "My",
			ServerSystemStore:       "LocalMachine",
			ServerTemplateID:        "1.2.3",
			ServerIssuerCommonNames: []string{"TEST"},
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "BAD_EXAMPLE", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	_, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)
	require.Error(t, err)
}

func TestTemplateIDFilter_Fail(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:       "My",
			ServerSystemStore: "LocalMachine",
			ServerTemplateID:  "1.2.3",
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2}, "", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	_, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)
	require.Error(t, err)
}

func TestMatching2CertsGetMostRecent(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ServerStore:       "My",
			ServerSystemStore: "LocalMachine",
			ServerTemplateID:  "1.2.3",
		},
	}
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-5)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "", "")
	shouldFind := makeCert(time.Now().Add(time.Duration(-1)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))
	serverSt.identities = append(serverSt.identities, newFakeIdentity(shouldFind))

	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	identity, err := c.findCertificate(c.cfg.ServerSystemStore, c.cfg.ServerStore, c.cfg.ServerIssuerCommonNames, c.cfg.ServerTemplateID, nil, findCert)

	require.NoError(t, err)
	foundCert, err := identity.Certificate()
	require.NoError(t, err)
	require.True(t, foundCert == shouldFind)
}

func TestRegularExpression(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ClientStore:        "My",
			ClientSystemStore:  "LocalMachine",
			ClientTemplateID:   "1.2.3",
			ClientSubjectRegEx: "[Villa]",
		},
	}
	var subjectRegEx *regexp.Regexp
	subjectRegEx, err := regexp.Compile(c.cfg.ClientSubjectRegEx)
	require.NoError(t, err)
	c.subjectRegEx = subjectRegEx
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-5)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "BobVilla", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))

	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	identity, err := c.findCertificate(c.cfg.ClientSystemStore, c.cfg.ClientStore, c.cfg.ClientIssuerCommonNames, c.cfg.ClientTemplateID, c.subjectRegEx, findCert)

	require.NoError(t, err)
	foundCert, err := identity.Certificate()
	require.NoError(t, err)
	require.True(t, foundCert == sc)
}

func TestRegularExpression_Fail(t *testing.T) {
	c := &winCertStoreHandler{
		cfg: WindowsCertificateFilter{
			ClientStore:        "My",
			ClientSystemStore:  "LocalMachine",
			ClientTemplateID:   "1.2.3",
			ClientSubjectRegEx: "[Villa]",
		},
	}
	var subjectRegEx *regexp.Regexp
	subjectRegEx, err := regexp.Compile(c.cfg.ClientSubjectRegEx)
	require.NoError(t, err)
	c.subjectRegEx = subjectRegEx
	serverSt := newFakeStore()
	sc := makeCert(time.Now().Add(time.Duration(-5)*time.Minute), time.Now().Add(5*time.Minute), []int{1, 2, 3}, "BAD_EXAMPLE", "")
	serverSt.identities = append(serverSt.identities, newFakeIdentity(sc))

	findCert := func(storeName, systemStore string) (certstore.Store, error) {
		if systemStore == "LocalMachine" {
			return serverSt, nil
		}
		return nil, fmt.Errorf("no store found")
	}
	identity, err := c.findCertificate(c.cfg.ClientSystemStore, c.cfg.ClientStore, c.cfg.ClientIssuerCommonNames, c.cfg.ClientTemplateID, c.subjectRegEx, findCert)

	require.Error(t, err)
	require.Nil(t, identity)

}

type fakeStore struct {
	identities []fakeIdentity
	closed     bool
}

func newFakeStore() fakeStore {
	return fakeStore{
		identities: make([]fakeIdentity, 0),
		closed:     false,
	}
}

func (f fakeStore) checkClosed(t *testing.T) {
	require.True(t, f.closed)
	for _, i := range f.identities {
		require.True(t, i.closed)
	}
}

func (f fakeStore) Identities() ([]certstore.Identity, error) {
	ids := make([]certstore.Identity, len(f.identities))
	for i, id := range f.identities {
		ids[i] = id
	}
	return ids, nil
}

func (f fakeStore) Import(data []byte, password string) error {
	panic("implement me")
}

func (f fakeStore) Close() {
	f.closed = true
}

var testAsnTemplateID = []int{1, 3, 6, 1, 4, 1, 311, 21, 7}

type fakeIdentity struct {
	cert   *x509.Certificate
	closed bool
}

func newFakeIdentity(cert *x509.Certificate) fakeIdentity {
	return fakeIdentity{cert: cert}
}

func (f fakeIdentity) Certificate() (*x509.Certificate, error) {
	return f.cert, nil
}

func makeCert(start, end time.Time, templateID []int, commonName string, subject string) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
			CommonName:    commonName,
		},
		NotBefore:             start,
		NotAfter:              end,
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		Issuer: pkix.Name{
			SerialNumber: subject,
			CommonName:   commonName,
		},
	}
	if len(templateID) != 0 {
		templateInfo := templateInformation{}
		templateInfo.Template = templateID
		ti, err := asn1.Marshal(templateInfo)
		if err != nil {
			println(err.Error())
			return nil
		}
		cert.Extensions = append(cert.Extensions, pkix.Extension{
			Id:       testAsnTemplateID,
			Critical: false,
			Value:    ti,
		})
	}
	return cert
}

func (f fakeIdentity) CertificateChain() ([]*x509.Certificate, error) {
	panic("implement me")
}

func (f fakeIdentity) Signer() (crypto.Signer, error) {
	panic("implement me")
}

func (f fakeIdentity) Delete() error {
	panic("implement me")
}

func (f fakeIdentity) Close() {
	f.closed = true
}
