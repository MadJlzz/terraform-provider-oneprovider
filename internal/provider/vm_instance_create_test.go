package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func readyInstance() *vm.InstanceReadResponse {
	info := &vm.InstanceReadResponse{}
	info.Response.ServerState.State = "online"
	info.Response.ServerInfo.IpAddress = "192.0.2.10"
	return info
}

func TestWaitForVMReadyRetriesTransientErrors(t *testing.T) {
	calls := 0
	err := waitForVMReady(context.Background(), 10*time.Second, vm.APIID("36"), func(context.Context, string) (*vm.InstanceReadResponse, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("client: api request failed with status: 502")
		}
		return readyInstance(), nil
	})
	if err != nil {
		t.Fatalf("waitForVMReady() error = %v, want nil after transient errors", err)
	}
	if calls != 3 {
		t.Fatalf("waitForVMReady() called the API getter %d times, want 3", calls)
	}
}

func TestWaitForVMReadyTimesOutWithLastError(t *testing.T) {
	err := waitForVMReady(context.Background(), time.Second, vm.APIID("36"), func(context.Context, string) (*vm.InstanceReadResponse, error) {
		return nil, errors.New("client: api request failed with status: 502")
	})
	if err == nil {
		t.Fatal("waitForVMReady() succeeded while the API kept failing")
	}
	if !strings.Contains(err.Error(), "status: 502") {
		t.Fatalf("waitForVMReady() error = %q, want it to contain the last API error", err)
	}
}

// fakeVMAPI is a minimal in-memory OneProvider VM API.
type fakeVMAPI struct {
	mu        sync.Mutex
	nextID    int
	ready     bool
	instances map[string]string // VM ID -> hostname
	destroyed []string
}

func newFakeVMAPI() *fakeVMAPI {
	return &fakeVMAPI{nextID: 1, instances: map[string]string{}}
}

func (f *fakeVMAPI) setReady(ready bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ready = ready
}

func (f *fakeVMAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	write := func(v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/vm/create":
		_ = r.ParseForm()
		id := strconv.Itoa(f.nextID)
		f.nextID++
		f.instances[id] = r.PostForm.Get("hostname")
		write(map[string]any{"response": map[string]any{
			"message":    "ok",
			"id":         id,
			"ip_address": "192.0.2.10",
			"hostname":   f.instances[id],
			"password":   "secret",
		}})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/vm/info/"):
		id := strings.TrimPrefix(r.URL.Path, "/vm/info/")
		hostname, ok := f.instances[id]
		if !ok {
			write(map[string]any{"error": map[string]any{"code": 810, "message": "not found"}})
			return
		}
		state := "online"
		if !f.ready {
			state = "offline"
		}
		write(map[string]any{"response": map[string]any{
			"server_install": !f.ready,
			"server_info":    map[string]any{"ipaddress": "192.0.2.10", "hostname": hostname},
			"server_state":   map[string]any{"status": "active", "state": state},
		}})
	case r.Method == http.MethodPost && r.URL.Path == "/vm/destroy":
		_ = r.ParseForm()
		id := r.PostForm.Get("vm_id")
		delete(f.instances, id)
		f.destroyed = append(f.destroyed, id)
		write(map[string]any{"response": map[string]any{"message": "ok"}})
	default:
		http.Error(w, "unexpected request "+r.Method+" "+r.URL.Path, http.StatusNotFound)
	}
}

func testVmInstanceResourceFakeAPIConfig(endpoint string) string {
	return fmt.Sprintf(`
provider "oneprovider" {
	endpoint = %q
}

resource "oneprovider_vm_instance" "test" {
	location_id      = "1"
	instance_size_id = "2"
	template_id      = "3"
	hostname         = "duplicate-check"

	timeouts {
		create = "2s"
	}
}
`, endpoint)
}

// Regression test for https://github.com/MadJlzz/terraform-provider-oneprovider/issues/77:
// a VM that fails to become ready must stay in state (tainted) and be replaced on the
// next apply, instead of being orphaned and duplicated.
func TestVmInstanceResource_readinessFailureKeepsVMInState(t *testing.T) {
	api := newFakeVMAPI()
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)

	t.Setenv(ApiKeyEnvVar, "test")
	t.Setenv(ClientKeyEnvVar, "test")

	config := testVmInstanceResourceFakeAPIConfig(server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			api.mu.Lock()
			defer api.mu.Unlock()
			if len(api.instances) != 0 {
				return fmt.Errorf("VMs left behind after destroy: %v", api.instances)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				// The VM never becomes ready within the create timeout.
				Config:      config,
				ExpectError: regexp.MustCompile("did not become ready"),
			},
			{
				// The failed VM is tainted in state, so it is replaced rather than duplicated.
				PreConfig: func() { api.setReady(true) },
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("oneprovider_vm_instance.test", plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("oneprovider_vm_instance.test", tfjsonpath.New("id"), knownvalue.StringExact("2")),
				},
				Check: func(*terraform.State) error {
					api.mu.Lock()
					defer api.mu.Unlock()
					if len(api.instances) != 1 {
						return fmt.Errorf("expected exactly one VM in the API, got %v", api.instances)
					}
					if len(api.destroyed) != 1 || api.destroyed[0] != "1" {
						return fmt.Errorf("expected the tainted VM 1 to be destroyed, got destroyed=%v", api.destroyed)
					}
					return nil
				},
			},
		},
	})
}
