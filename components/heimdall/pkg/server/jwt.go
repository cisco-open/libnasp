// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
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

package server

import (
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/golang-jwt/jwt/v4"
	authv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	naspTokenAudiencePrefix = "nasp-heimdall:workloadgroup:" //nolint:gosec
	naspTokenAudienceFormat = naspTokenAudiencePrefix + "%s:%s"
)

type NASPTokenAudience string

func (a NASPTokenAudience) Generate(ref client.ObjectKey) string {
	return fmt.Sprintf(naspTokenAudienceFormat, ref.Namespace, ref.Name)
}

func (a NASPTokenAudience) Valid() bool {
	return strings.HasPrefix(string(a), naspTokenAudiencePrefix)
}

func (a NASPTokenAudience) String() string {
	return string(a)
}

func (a NASPTokenAudience) Parse() client.ObjectKey {
	ref := client.ObjectKey{}

	if p := strings.Split(string(a), ":"); len(p) == 4 {
		ref.Name = p[3]
		ref.Namespace = p[2]
	}

	return ref
}

type AuthenticatedUser struct {
	UID               string
	Name              string
	Groups            []string
	ServiceAccountRef client.ObjectKey
	WorkloadGroupRef  client.ObjectKey
}

func AuthenticateToken(ctx context.Context, c client.Client, token string) (*AuthenticatedUser, error) {
	var naspAudience NASPTokenAudience

	// parse JWT token to extract NASP audience
	claims := jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, &claims)
	if err != nil {
		return nil, errors.WrapIf(err, "could not parse JWT token")
	}
	for _, aud := range claims.Audience {
		if naspAud := NASPTokenAudience(aud); naspAud.Valid() {
			naspAudience = naspAud
			break
		}
	}

	// create token review
	tokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{Token: token, Audiences: []string{naspAudience.String()}},
	}
	err = c.Create(ctx, tokenReview)
	if err != nil {
		return nil, errors.WrapIf(err, "could not create token review")
	}

	if tokenReview.Status.Error != "" {
		return nil, errors.New(tokenReview.Status.Error)
	}

	if !tokenReview.Status.Authenticated {
		return nil, errors.New("token authentication failed")
	}

	user := &AuthenticatedUser{
		UID:              tokenReview.Status.User.UID,
		Name:             tokenReview.Status.User.Username,
		Groups:           tokenReview.Status.User.Groups,
		WorkloadGroupRef: naspAudience.Parse(),
	}

	// extract service account info from username
	if strings.HasPrefix(tokenReview.Status.User.Username, "system:serviceaccount:") {
		// The username is of format: system:serviceaccount:(NAMESPACE):(SERVICEACCOUNT)
		parts := strings.Split(tokenReview.Status.User.Username, ":")
		if len(parts) != 4 {
			return nil, errors.New("username format is invalid")
		}

		user.ServiceAccountRef = client.ObjectKey{
			Name:      parts[3],
			Namespace: parts[2],
		}
	}

	return user, nil
}
