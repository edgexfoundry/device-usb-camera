// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

// Credentials encapsulates Username and Password pair.
type Credentials struct {
	Username string
	Password string
}

// tryGetCredentials will attempt one time to get the credentials located at secretPath from
// secret provider and return them, otherwise return an error.
func (d *Driver) tryGetCredentials(secretPath string) (Credentials, errors.EdgeX) {
	secretData, err := d.ds.SecretProvider().GetSecret(secretPath, secret.UsernameKey, secret.PasswordKey)
	if err != nil {
		d.lc.Errorf("Failed to retrieve credentials for the secret path %s: %s", secretPath, err)
		return Credentials{}, errors.NewCommonEdgeXWrapper(err)
	}

	credentials := Credentials{
		Username: secretData[secret.UsernameKey],
		Password: secretData[secret.PasswordKey],
	}

	return credentials, nil
}
