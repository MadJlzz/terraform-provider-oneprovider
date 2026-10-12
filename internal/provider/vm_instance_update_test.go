package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
)

func instanceWithHostname(hostname string) *vm.InstanceReadResponse {
	info := readyInstance()
	info.Response.ServerInfo.Hostname = hostname
	return info
}

func hostnameUpdateRequest() *vm.InstanceHostnameUpdateRequest {
	return &vm.InstanceHostnameUpdateRequest{VmId: "36", Hostname: "new-hostname"}
}

func TestUpdateVMHostnameRetriesWhileLockedAndWaitsForHostname(t *testing.T) {
	updateCalls, getCalls := 0, 0
	err := updateVMHostname(context.Background(), 10*time.Second, hostnameUpdateRequest(),
		func(context.Context, *vm.InstanceHostnameUpdateRequest) error {
			updateCalls++
			if updateCalls < 3 {
				return apiError(apiErrOperationFailed, "VPS Locked")
			}
			return nil
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			getCalls++
			if getCalls < 2 {
				return instanceWithHostname("old-hostname"), nil
			}
			return instanceWithHostname("new-hostname"), nil
		},
	)
	if err != nil {
		t.Fatalf("updateVMHostname() error = %v, want nil", err)
	}
	if updateCalls != 3 {
		t.Fatalf("updateVMHostname() called update %d times, want 3", updateCalls)
	}
	if getCalls != 2 {
		t.Fatalf("updateVMHostname() called the API getter %d times, want 2", getCalls)
	}
}

func TestUpdateVMHostnameFailsFastOnOtherErrors(t *testing.T) {
	updateCalls := 0
	err := updateVMHostname(context.Background(), 10*time.Second, hostnameUpdateRequest(),
		func(context.Context, *vm.InstanceHostnameUpdateRequest) error {
			updateCalls++
			return apiError(apiErrOperationFailed, "Operation not performed")
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			t.Fatal("updateVMHostname() polled the VM after a failed update")
			return nil, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "Operation not performed") {
		t.Fatalf("updateVMHostname() error = %v, want the API error", err)
	}
	if updateCalls != 1 {
		t.Fatalf("updateVMHostname() called update %d times, want 1", updateCalls)
	}
}

func TestUpdateVMHostnameTimesOutWhileLocked(t *testing.T) {
	err := updateVMHostname(context.Background(), time.Second, hostnameUpdateRequest(),
		func(context.Context, *vm.InstanceHostnameUpdateRequest) error {
			return apiError(apiErrOperationFailed, "VPS Locked")
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			t.Fatal("updateVMHostname() polled the VM while it was still locked")
			return nil, nil
		},
	)
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != apiErrOperationFailed {
		t.Fatalf("updateVMHostname() error = %v, want the last VPS Locked error", err)
	}
}

func TestUpdateVMHostnameReportsUnconfirmedUpdate(t *testing.T) {
	err := updateVMHostname(context.Background(), time.Second, hostnameUpdateRequest(),
		func(context.Context, *vm.InstanceHostnameUpdateRequest) error { return nil },
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			return instanceWithHostname("old-hostname"), nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "accepted but could not be confirmed") {
		t.Fatalf("updateVMHostname() error = %v, want an unconfirmed update error", err)
	}
}
