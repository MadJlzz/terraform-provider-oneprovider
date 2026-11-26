package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestVmInstanceResource_tags(t *testing.T) {
	api := newFakeVMAPI()
	api.setReady(true)
	// 101 is also assigned to another service; its assignment must never be touched.
	assignees := map[int][]int{101: {2}, 102: {}, 2000: {}}
	var changes []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/tag/") {
			api.ServeHTTP(w, r)
			return
		}
		api.mu.Lock()
		defer api.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tag/listing/"):
			project := strings.TrimPrefix(r.URL.Path, "/tag/listing/")
			ids := []int{101, 102}
			if project == testProdProjectID {
				ids = []int{2000}
			} else if project != "default-test-id" {
				http.NotFound(w, r)
				return
			}
			tags := []map[string]any{}
			for _, id := range ids {
				label := "terraform managed"
				if id == 102 {
					label = "backup"
				}
				tags = append(tags, map[string]any{"id": id, "label": label, "project_uuid": project, "assignees": assignees[id]})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": "success", "response": tags})
		case r.Method == http.MethodPost && (r.URL.Path == "/tag/assign/" || r.URL.Path == "/tag/unassign/"):
			id, err := strconv.Atoi(r.PostFormValue("tag_id"))
			if _, ok := assignees[id]; !ok || err != nil || len(r.PostForm) != 2 || r.PostFormValue("service_id") != "1" {
				http.Error(w, "unexpected tag request", http.StatusBadRequest)
				return
			}
			changes = append(changes, fmt.Sprintf("%s%d", r.URL.Path, id))
			if r.URL.Path == "/tag/assign/" {
				assignees[id] = append(assignees[id], 1)
			} else {
				assignees[id] = slices.DeleteFunc(assignees[id], func(i int) bool { return i == 1 })
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")

	config := func(project, tags string) string {
		projectAttr := ""
		if project != "" {
			projectAttr = fmt.Sprintf("project = %q", project)
		}
		return fmt.Sprintf(`
provider "oneprovider" { endpoint = %q }
data "oneprovider_vm_tag" "managed" {
  name = "terraform managed"
  %[2]s
}
resource "oneprovider_vm_instance" "test" {
  location_id = "1"
  instance_size_id = "2"
  template_id = "3"
  hostname = "tag-test"
  %[2]s
  %[3]s
}
`, server.URL, projectAttr, tags)
	}
	check := func(want ...string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("oneprovider_vm_instance.test", "id", "1"), func(*terraform.State) error {
			api.mu.Lock()
			defer api.mu.Unlock()
			if !slices.Equal(changes, want) || !slices.Contains(assignees[101], 2) {
				return fmt.Errorf("unexpected tag changes %v, assignments %v", changes, assignees)
			}
			changes = nil
			return nil
		})
	}
	const managed = `tags = [data.oneprovider_vm_tag.managed.id]`
	const both = `tags = ["102", data.oneprovider_vm_tag.managed.id]`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("", managed), Check: check("/tag/assign/101")},
			{Config: config("", both), Check: check("/tag/assign/102")},
			{Config: config("", `tags = ["102", "9999"]`), ExpectError: regexp.MustCompile(`tag "9999" not found`)},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					assignees[101] = []int{2} // simulate removal in the panel
				},
				Config: config("", both),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("oneprovider_vm_instance.test", plancheck.ResourceActionUpdate),
				}},
				Check: check("/tag/assign/101"),
			},
			{Config: config("", managed), Check: check("/tag/unassign/102")},
			{Config: config(testProdProjectID, managed), Check: check("/tag/unassign/101", "/tag/assign/2000")},
			{Config: config(testProdProjectID, ""), Check: check("/tag/unassign/2000")},
			{Config: config(testProdProjectID, ""), Check: check()},
		},
	})
}
