/**
 * @Author: alessonhu
 * @Description:
 * @File:  command.go
 * @Version: 1.0.0
 * @Date: 2021/1/12 20:09
 */
package main

import (
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/groob/plist"
	"mdm/common/qm"
)

const (
	CommandInstallProfile = "InstallProfile"
)

type Command struct {
	Command     requestType
	CommandUUID string
}

type requestType struct {
	// InstallProfile
	RequestType string `plist:",omitempty"`
	Payload     []byte `plist:",omitempty"`
}

func NewCommandPayload(payload []byte) (pUuid string, profile []byte, err error) {
	pUuid = uuid.New().String()
	payloadData, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		qm.LOG_ERROR(err)
		return
	}
	command := Command{
		CommandUUID: pUuid,
		Command: requestType{
			RequestType: CommandInstallProfile,
			Payload:     payloadData,
		},
	}
	profile, err = plist.MarshalIndent(command, "\t")
	if err != nil {
		qm.LOG_ERROR(err)
		return
	}
	return
}
