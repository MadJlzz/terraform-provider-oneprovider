package provider

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"golang.org/x/crypto/ssh"
)

// testAccNamePrefix is prepended to every resource name created by acceptance tests,
// making leftovers of interrupted runs easy to spot and clean up.
const testAccNamePrefix = "tfacc"

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"oneprovider": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if _, ok := os.LookupEnv(ApiKeyEnvVar); !ok {
		t.Fatalf("missing %s environment variable", ApiKeyEnvVar)
	}
	if _, ok := os.LookupEnv(ClientKeyEnvVar); !ok {
		t.Fatalf("missing %s environment variable", ClientKeyEnvVar)
	}
}

// testAccRandomName returns a unique lowercase alphanumeric name, so that concurrent
// test runs (e.g. parallel CI pipelines) do not conflict with each other.
// It satisfies the SSH key name constraint (^[a-z0-9]*$) and is a valid hostname.
func testAccRandomName() string {
	return testAccNamePrefix + acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
}

// testAccRandomSSHPublicKey generates a fresh ed25519 public key in OpenSSH authorized
// key format. OneProvider enforces uniqueness on the key material itself (not on the
// name), so each test run needs its own key to avoid conflicts.
func testAccRandomSSHPublicKey(t *testing.T) string {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to convert ed25519 key to SSH public key: %v", err)
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
}
