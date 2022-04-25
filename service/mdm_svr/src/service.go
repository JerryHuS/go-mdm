/**
 * @Author: alessonhu
 * @Description:
 * @File:  service.go
 * @Version: 1.0.0
 * @Date: 2020/12/28 20:38
 */
package main

import (
	"bytes"
	"crypto/x509"
	"github.com/groob/plist"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	EnrollmentProfileId string = "ioa.mdm.enroll"
	OTAProfileId        string = "ioa.mdm.ota"
	PerUserConnections         = "com.apple.mdm.per-user-connections"
)

type mdmService struct {
	ORGANIZATION  string
	SVRURL        string
	SCEPURL       string
	SCEPChallenge string
	SCEPSubject   [][][]string
	TLSCert       []byte
	TLSPass       string

	mu sync.RWMutex
	//Topic string // APNS Topic for MDM notifications
}

func (svc *mdmService) MakeEnrollmentProfile(mid string) (Profile, error) {
	profile := NewProfile()
	profile.PayloadIdentifier = EnrollmentProfileId
	profile.PayloadOrganization = svc.ORGANIZATION
	profile.PayloadDisplayName = "MDMEnrollProfile"
	profile.PayloadDescription = "The server may alter your settings"
	profile.PayloadScope = "System"

	mdmPayload := NewPayload("com.apple.mdm")
	mdmPayload.PayloadDescription = "Enrolls with the MDM server"
	mdmPayload.PayloadOrganization = svc.ORGANIZATION
	mdmPayload.PayloadIdentifier = EnrollmentProfileId + ".mdm"
	mdmPayload.PayloadScope = "System"

	svc.mu.Lock()
	topic := "com.apple.mgmt.External." + g_app.ApnsSvc.PushCertTopic
	svc.mu.Unlock()

	mdmPayloadContent := MDMPayloadContent{
		Payload:             *mdmPayload,
		AccessRights:        allRights(),
		CheckInURL:          svc.SVRURL + "/api/public/mdm/checkin?mid=" + mid,
		CheckOutWhenRemoved: true,
		ServerURL:           svc.SVRURL + "/api/public/mdm/connect",
		Topic:               topic,
		SignMessage:         true,
		ServerCapabilities:  []string{PerUserConnections},
	}

	payloadContent := []interface{}{}

	if svc.SCEPURL != "" {
		scepContent := SCEPPayloadContent{
			URL:      svc.SCEPURL,
			Keysize:  2048,
			KeyType:  "RSA",
			KeyUsage: int(x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment),
			Name:     "Device Management Identity Certificate",
			Subject:  svc.SCEPSubject,
		}

		scepPayload := NewPayload("com.apple.security.scep")
		scepPayload.PayloadDescription = "Configures SCEP"
		scepPayload.PayloadDisplayName = "SCEP"
		scepPayload.PayloadIdentifier = EnrollmentProfileId + ".scep"
		scepPayload.PayloadOrganization = svc.ORGANIZATION
		scepPayload.PayloadContent = scepContent
		scepPayload.PayloadScope = "System"

		mdmPayloadContent.IdentityCertificateUUID = scepPayload.PayloadUUID

		payloadContent = append(payloadContent, *scepPayload)
	}

	// Client needs to trust us at this point if we are using a self signed certificate.
	if len(svc.TLSCert) > 0 {
		tlsPayload := NewPayload("com.apple.security.pkcs12")
		tlsPayload.Password = svc.TLSPass
		tlsPayload.PayloadDisplayName = "TLS certificate for MDM"
		tlsPayload.PayloadDescription = "Installs the TLS certificate for MDM"
		tlsPayload.PayloadIdentifier = EnrollmentProfileId + ".cert.selfsigned"
		tlsPayload.PayloadContent = svc.TLSCert
		tlsPayload.PayloadOrganization = svc.ORGANIZATION

		mdmPayloadContent.IdentityCertificateUUID = tlsPayload.PayloadUUID

		payloadContent = append(payloadContent, mdmPayloadContent)
		payloadContent = append(payloadContent, *tlsPayload)
	}

	// 描述文件密码
	passwordPayload := NewPayload("com.apple.profileRemovalPassword")
	passwordContent := PasswordPayloadContent{
		Payload:         *passwordPayload,
		PayloadEnabled:  true,
		RemovalPassword: GetRandomString(6),
	}
	payloadContent = append(payloadContent, passwordContent)

	profile.PayloadContent = payloadContent

	return *profile, nil
}

func profileOrPayloadFromFunc(f interface{}) (interface{}, error) {
	fPayload, ok := f.(func() (Payload, error))
	if !ok {
		fProfile := f.(func() (Profile, error))
		return fProfile()
	}
	return fPayload()
}

func profileOrPayloadToMobileconfig(in interface{}) (Mobileconfig, error) {
	if _, ok := in.(Payload); !ok {
		_ = in.(Profile)
	}
	buf := new(bytes.Buffer)
	enc := plist.NewEncoder(buf)
	enc.Indent("  ")
	err := enc.Encode(in)
	return buf.Bytes(), err
}

func rspMdmPlist() (res []byte) {
	res, err := plist.MarshalIndent(struct{}{}, "\t")
	if err != nil {
		return nil
	}
	return
}

// Extract (raw) body bytes, parse property list
func mdmRequestBody(r *http.Request, s interface{}) ([]byte, error) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = plist.Unmarshal(body, s)
	if err != nil {
		return body, err
	}

	return body, nil
}
