package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Fixed identity provider config the test realm (testdata/evently-realm-export.json)
// is built around. There's no official testcontainers-go Keycloak module (unlike
// Postgres/Redis), so the container is assembled by hand from a generic image.
const (
	keycloakRealm                    = "evently"
	keycloakPublicClientID           = "evently-public-client"
	keycloakConfidentialClientID     = "evently-confidential-client"
	keycloakConfidentialClientSecret = "test-confidential-secret"
)

type keycloakContainer struct {
	testcontainers.Container
	issuerURL string
	adminURL  string
	tokenURL  string
}

func startKeycloak(ctx context.Context) (*keycloakContainer, error) {
	ctr, err := testcontainers.Run(ctx, "quay.io/keycloak/keycloak:latest",
		testcontainers.WithEnv(map[string]string{
			"KEYCLOAK_ADMIN":          "admin",
			"KEYCLOAK_ADMIN_PASSWORD": "admin",
		}),
		testcontainers.WithCmd("start-dev", "--import-realm"),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      "testdata/evently-realm-export.json",
			ContainerFilePath: "/opt/keycloak/data/import/realm.json",
			FileMode:          0o644,
		}),
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/realms/"+keycloakRealm+"/.well-known/openid-configuration").
				WithPort("8080/tcp").
				WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start keycloak container: %w", err)
	}

	endpoint, err := ctr.PortEndpoint(ctx, "8080/tcp", "http")
	if err != nil {
		return nil, fmt.Errorf("keycloak endpoint: %w", err)
	}

	return &keycloakContainer{
		Container: ctr,
		issuerURL: fmt.Sprintf("%s/realms/%s", endpoint, keycloakRealm),
		adminURL:  fmt.Sprintf("%s/admin/realms/%s/", endpoint, keycloakRealm),
		tokenURL:  fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", endpoint, keycloakRealm),
	}, nil
}
