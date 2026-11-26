package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestVmTagResource(t *testing.T) {
	const address = "oneprovider_vm_tag.test"
	type tag struct {
		ID          int    `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
		Project     string `json:"project_uuid"`
	}
	var mu sync.Mutex
	other := tag{ID: 101, Label: "test", Project: testDevProjectID}
	stored := map[int]tag{other.ID: other}
	nextID := 102
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		write := func(v any) { _ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "response": v}) }
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tag/listing/") {
			list := []tag{}
			for _, item := range stored {
				if item.Project == strings.TrimPrefix(r.URL.Path, "/tag/listing/") {
					list = append(list, item)
				}
			}
			write(list)
			return
		}
		if r.Method != http.MethodPost || r.ParseForm() != nil {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		form := r.PostForm
		id, _ := strconv.Atoi(form.Get("tag_id"))
		switch r.URL.Path {
		case "/tag/add":
			if len(form) == 3 && form.Get("project_uuid") != "" && form.Get("label") != "" && form.Has("description") {
				item := tag{nextID, form.Get("label"), form.Get("description"), form.Get("project_uuid")}
				stored[nextID] = item
				nextID++
				write(item)
				return
			}
		case "/tag/edit":
			if item, ok := stored[id]; ok && len(form) == 3 && form.Get("label") != "" && form.Has("description") {
				item.Label, item.Description = form.Get("label"), form.Get("description")
				stored[id] = item
				write(item)
				return
			}
		case "/tag/remove":
			if _, ok := stored[id]; ok && len(form) == 1 {
				delete(stored, id)
				write("ok")
				return
			}
		}
		http.Error(w, "unexpected form or endpoint", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")
	config := func(project, label, descriptionAttr string) string {
		return fmt.Sprintf(`
provider "oneprovider" { endpoint = %q }
resource "oneprovider_vm_tag" "test" {
  project = %q
  label = %q
  %s
}
`, server.URL, project, label, descriptionAttr)
	}
	check := func(id, project, label, description string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(address, "id", id),
			resource.TestCheckResourceAttr(address, "project", project),
			resource.TestCheckResourceAttr(address, "label", label),
			resource.TestCheckResourceAttr(address, "description", description),
		)
	}
	plan := func(action plancheck.ResourceActionType) resource.ConfigPlanChecks {
		return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, action)}}
	}
	dev := config(testDevProjectID, "renamed", "")
	prod := config(testProdProjectID, "renamed", "")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(testDevProjectID, "test", `description = "test desc"`), Check: check("102", testDevProjectID, "test", "test desc")},
			{Config: dev, ConfigPlanChecks: plan(plancheck.ResourceActionUpdate), Check: check("102", testDevProjectID, "renamed", "")},
			{ResourceName: address, ImportState: true, ImportStateId: testDevProjectID + "/102", ImportStateVerify: true},
			{Config: prod, ConfigPlanChecks: plan(plancheck.ResourceActionReplace), Check: check("103", testProdProjectID, "renamed", "")},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					delete(stored, 103)
				},
				Config: prod, ConfigPlanChecks: plan(plancheck.ResourceActionCreate), Check: check("104", testProdProjectID, "renamed", ""),
			},
		},
		CheckDestroy: func(*terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if len(stored) != 1 || stored[other.ID] != other {
				return fmt.Errorf("unexpected tags remaining after destroy: %v", stored)
			}
			return nil
		},
	})
}

func TestVmTagResource_createMissingID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tag/add" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `{"result":"success","response":{"label":"test"}}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`provider "oneprovider" { endpoint = %q }
resource "oneprovider_vm_tag" "test" {
  project = "project-1"
  label = "test"
}`, server.URL),
			ExpectError: regexp.MustCompile("empty tag ID"),
		}},
	})
}
