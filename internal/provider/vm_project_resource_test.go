package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestVmProjectResource(t *testing.T) {
	const address = "oneprovider_vm_project.test"
	type project struct {
		UUID      string `json:"uuid"`
		Name      string `json:"name"`
		IsDefault int    `json:"is_default"`
	}
	var mu sync.Mutex
	other := project{UUID: "unmanaged", Name: "example"}
	stored := map[string]project{other.UUID: other}
	nextID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		write := func(v any) { _ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "response": v}) }
		if r.Method == http.MethodGet && r.URL.Path == "/vm/project/list" {
			list := []project{}
			for _, p := range stored {
				list = append(list, p)
			}
			write(map[string]any{"projects": list})
			return
		}
		if r.Method != http.MethodPost || r.ParseForm() != nil {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		form := r.PostForm
		id := form.Get("project_uuid")
		flag, err := strconv.Atoi(form.Get("is_default"))
		validFields := err == nil && (flag == 0 || flag == 1) && form.Get("name") != ""
		switch r.URL.Path {
		case "/vm/project/new":
			if len(form) == 2 && validFields {
				id = fmt.Sprintf("project-%d", nextID)
				nextID++
				stored[id] = project{id, form.Get("name"), flag}
				write(map[string]any{"project": stored[id]})
				return
			}
		case "/vm/project/edit":
			if _, ok := stored[id]; ok && len(form) == 3 && validFields {
				stored[id] = project{id, form.Get("name"), flag}
				write("ok")
				return
			}
		case "/vm/project/delete":
			if _, ok := stored[id]; ok && len(form) == 1 {
				delete(stored, id)
				write("ok")
				return
			}
		}
		http.Error(w, "unexpected endpoint or form", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")
	config := func(name, defaultAttr string) string {
		return fmt.Sprintf(`provider "oneprovider" { endpoint = %q }
resource "oneprovider_vm_project" "test" {
  name = %q
  %s
}`, server.URL, name, defaultAttr)
	}
	check := func(id, name, flag string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(address, "id", id),
			resource.TestCheckResourceAttr(address, "name", name),
			resource.TestCheckResourceAttr(address, "is_default", flag),
		)
	}
	updated := config("renamed", "is_default = true")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("example", ""), Check: check("project-1", "example", "false")},
			{Config: updated, Check: check("project-1", "renamed", "true"), ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
			}},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					delete(stored, "project-1")
				},
				Config: updated, Check: check("project-2", "renamed", "true"), ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate)},
				},
			},
		},
		CheckDestroy: func(*terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if len(stored) != 1 || stored[other.UUID] != other {
				return fmt.Errorf("unexpected projects after destroy: %v", stored)
			}
			return nil
		},
	})
}

func TestVmProjectResource_createMissingUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vm/project/new" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `{"result":"success","response":{"project":{"name":"example","is_default":0}}}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`provider "oneprovider" { endpoint = %q }
resource "oneprovider_vm_project" "test" { name = "example" }`, server.URL),
			ExpectError: regexp.MustCompile("empty project UUID"),
		}},
	})
}
