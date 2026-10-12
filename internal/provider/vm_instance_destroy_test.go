package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
)

func apiError(code int, message string) error {
	return fmt.Errorf("vm: destroy instance failed: %w", &client.APIError{Code: code, Message: message})
}

func TestDestroyVMRetriesWhileLockedAndWaitsUntilGone(t *testing.T) {
	destroyCalls, getCalls := 0, 0
	err := destroyVM(context.Background(), 10*time.Second, "36",
		func(context.Context, *vm.InstanceDestroyRequest) error {
			destroyCalls++
			if destroyCalls < 3 {
				return apiError(apiErrVMLocked, "VPS Locked")
			}
			return nil
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			getCalls++
			if getCalls < 2 {
				return readyInstance(), nil
			}
			return nil, apiError(apiErrVMNotFound, "not found")
		},
	)
	if err != nil {
		t.Fatalf("destroyVM() error = %v, want nil", err)
	}
	if destroyCalls != 3 {
		t.Fatalf("destroyVM() called destroy %d times, want 3", destroyCalls)
	}
	if getCalls != 2 {
		t.Fatalf("destroyVM() called the API getter %d times, want 2", getCalls)
	}
}

func TestDestroyVMFailsFastOnOtherErrors(t *testing.T) {
	destroyCalls := 0
	err := destroyVM(context.Background(), 10*time.Second, "36",
		func(context.Context, *vm.InstanceDestroyRequest) error {
			destroyCalls++
			return apiError(1, "permission denied")
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			t.Fatal("destroyVM() polled the VM after a failed destroy")
			return nil, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("destroyVM() error = %v, want the API error", err)
	}
	if destroyCalls != 1 {
		t.Fatalf("destroyVM() called destroy %d times, want 1", destroyCalls)
	}
}

func TestDestroyVMTreatsMissingVMAsDestroyed(t *testing.T) {
	err := destroyVM(context.Background(), 10*time.Second, "36",
		func(context.Context, *vm.InstanceDestroyRequest) error {
			return apiError(apiErrVMNotFound, "not found")
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			return nil, apiError(apiErrVMNotFound, "not found")
		},
	)
	if err != nil {
		t.Fatalf("destroyVM() error = %v, want nil", err)
	}
}

func TestDestroyVMTimesOutWhileLocked(t *testing.T) {
	err := destroyVM(context.Background(), time.Second, "36",
		func(context.Context, *vm.InstanceDestroyRequest) error {
			return apiError(apiErrVMLocked, "VPS Locked")
		},
		func(context.Context, string) (*vm.InstanceReadResponse, error) {
			t.Fatal("destroyVM() polled the VM while it was still locked")
			return nil, nil
		},
	)
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != apiErrVMLocked {
		t.Fatalf("destroyVM() error = %v, want the last VPS Locked error", err)
	}
}
