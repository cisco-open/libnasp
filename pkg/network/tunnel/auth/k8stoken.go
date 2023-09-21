// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package auth

import (
	"context"
	"strings"

	"emperror.dev/errors"
	authv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cisco-open/libnasp/pkg/network/tunnel/api"
)

var (
	DefaultK8sAuthenticatorAud = "https://kubernetes.default.svc.cluster.local"
)

type k8sAuthenticator struct {
	c   client.Client
	aud []string
}

type K8sAuthenticatorOption func(*k8sAuthenticator)

func K8sAuthenticatorWithAud(aud ...string) K8sAuthenticatorOption {
	return func(a *k8sAuthenticator) {
		a.aud = aud
	}
}

func NewK8sAuthenticator(c client.Client, options ...K8sAuthenticatorOption) api.Authenticator {
	a := &k8sAuthenticator{
		c:   c,
		aud: []string{DefaultK8sAuthenticatorAud},
	}

	for _, option := range options {
		option(a)
	}

	return a
}

func (a *k8sAuthenticator) Authenticate(ctx context.Context, token string) (bool, api.User, error) {
	tokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{Token: token, Audiences: a.aud},
	}

	err := a.c.Create(ctx, tokenReview)
	if err != nil {
		return false, nil, errors.WrapIf(err, "could not create token review")
	}

	if tokenReview.Status.Error != "" {
		return false, nil, errors.New(tokenReview.Status.Error)
	}

	if !tokenReview.Status.Authenticated {
		return false, nil, errors.New("token authentication failed")
	}

	user := &k8sUser{
		uid:    tokenReview.Status.User.UID,
		name:   tokenReview.Status.User.Username,
		groups: tokenReview.Status.User.Groups,
	}

	if l := strings.Split(user.Name(), ":"); len(l) == 4 && l[0] == "system" {
		user.metadata = map[string]string{
			"namespace": l[2],
			l[1]:        l[3],
		}
	}

	return true, user, nil
}
