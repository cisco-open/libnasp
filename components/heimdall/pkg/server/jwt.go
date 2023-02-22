package server

import (
	"context"
	"emperror.dev/errors"
	authv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type AuthenticatedUser struct {
	Name           string
	Namespace      string
	ServiceAccount string
	Groups         []string
}

func AuthenticateToken(ctx context.Context, client client.Client, token string) (*AuthenticatedUser, error) {
	tokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{Token: token},
	}
	err := client.Create(ctx, tokenReview)
	if err != nil {
		return nil, err
	}

	if tokenReview.Status.Error != "" {
		return nil, errors.New("invalid token")
	}

	if !tokenReview.Status.Authenticated {
		return nil, errors.New("invalid token")
	}

	if strings.HasPrefix(tokenReview.Status.User.Username, "system:serviceaccount:") {
		// The username is of format: system:serviceaccount:(NAMESPACE):(SERVICEACCOUNT)
		parts := strings.Split(tokenReview.Status.User.Username, ":")
		if len(parts) != 4 {
			return nil, errors.New("username format is invalid")
		}

		return &AuthenticatedUser{
			Name:           parts[3],
			Namespace:      parts[2],
			ServiceAccount: tokenReview.Status.User.Username,
			Groups:         tokenReview.Status.User.Groups,
		}, nil
	}

	return &AuthenticatedUser{
		Name:   tokenReview.Status.User.Username,
		Groups: tokenReview.Status.User.Groups,
	}, nil
}
