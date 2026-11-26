package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVmProjectDataSource_lookupByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/vm/project/list" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"response":{"projects":[
			{"name":"shared","uuid":"uuid-a"},
			{"name":"shared","uuid":"uuid-b"},
			{"name":"invalid","uuid":""}
		]}}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")

	config := func(name string) string {
		return fmt.Sprintf(`
provider "oneprovider" { endpoint = %q }
data "oneprovider_vm_project" "test" { name = %q }
`, server.URL, name)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(""), ExpectError: regexp.MustCompile("provide a non-empty project name")},
			{Config: config("missing"), ExpectError: regexp.MustCompile("project not found for name missing")},
			{
				Config: config("shared"),
				Check:  resource.TestCheckResourceAttr("data.oneprovider_vm_project.test", "id", "uuid-a"),
			},
			{Config: config("invalid"), ExpectError: regexp.MustCompile("API returned an empty project UUID")},
		},
	})
}
