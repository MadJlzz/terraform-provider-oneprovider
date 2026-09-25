package vm

import (
	"encoding/json"
	"testing"
)

func TestAPIID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{name: "number", payload: `36`, want: "36"},
		{name: "string", payload: `"36"`, want: "36"},
		{name: "exact large number", payload: `9007199254740993`, want: "9007199254740993"},
		{name: "quoted leading zeros", payload: `"0036"`, want: "0036"},
		{name: "opaque string", payload: `"vm-36"`, want: "vm-36"},
		{name: "null", payload: `null`, wantErr: true},
		{name: "boolean", payload: `true`, wantErr: true},
		{name: "object", payload: `{}`, wantErr: true},
		{name: "array", payload: `[]`, wantErr: true},
		{name: "empty string", payload: `""`, wantErr: true},
		{name: "negative number", payload: `-36`, wantErr: true},
		{name: "fractional number", payload: `36.5`, wantErr: true},
		{name: "exponent number", payload: `36e0`, wantErr: true},
		{name: "malformed JSON", payload: `36x`, wantErr: true},
		{name: "trailing JSON", payload: `36 37`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got APIID
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s, got APIID %q", tt.payload, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("String() = %q, want %q", got.String(), tt.want)
			}
			if string(got) != tt.want {
				t.Errorf("APIID = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocationsListResponseID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "number", payload: `{"response":{"region":[{"id":36}]}}`, want: "36"},
		{name: "string", payload: `{"response":{"region":[{"id":"36"}]}}`, want: "36"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got LocationsListResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			locations := got.Response["region"]
			if len(locations) != 1 {
				t.Fatalf("got %d locations, want 1", len(locations))
			}
			if locations[0].Id.String() != tt.want {
				t.Errorf("ID = %q, want %q", locations[0].Id, tt.want)
			}
		})
	}
}

func TestSizesListResponseID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "number", payload: `{"response":[{"id":36}]}`, want: "36"},
		{name: "string", payload: `{"response":[{"id":"36"}]}`, want: "36"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SizesListResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Response) != 1 {
				t.Fatalf("got %d sizes, want 1", len(got.Response))
			}
			if got.Response[0].Id.String() != tt.want {
				t.Errorf("ID = %q, want %q", got.Response[0].Id, tt.want)
			}
		})
	}
}

func TestTemplatesListResponseID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "number", payload: `{"response":[{"id":36}]}`, want: "36"},
		{name: "string", payload: `{"response":[{"id":"36"}]}`, want: "36"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TemplatesListResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Templates) != 1 {
				t.Fatalf("got %d templates, want 1", len(got.Templates))
			}
			if got.Templates[0].Id.String() != tt.want {
				t.Errorf("ID = %q, want %q", got.Templates[0].Id, tt.want)
			}
		})
	}
}

func TestInstanceCreateResponseID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{name: "number", payload: `{"response":{"id":36}}`, want: "36"},
		{name: "string", payload: `{"response":{"id":"36"}}`, want: "36"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got InstanceCreateResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Response.Id.String() != tt.want {
				t.Errorf("ID = %q, want %q", got.Response.Id, tt.want)
			}
		})
	}
}

func TestNumericString(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{name: "number", payload: `36`, want: "36"},
		{name: "numeric string", payload: `"36"`, want: "36"},
		{name: "exact large integer", payload: `9007199254740993`, want: "9007199254740993"},
		{name: "quoted leading zeros", payload: `"0036"`, want: "0036"},
		{name: "null", payload: `null`, wantErr: true},
		{name: "boolean", payload: `true`, wantErr: true},
		{name: "object", payload: `{}`, wantErr: true},
		{name: "array", payload: `[]`, wantErr: true},
		{name: "empty string", payload: `""`, wantErr: true},
		{name: "non-numeric string", payload: `"cores"`, wantErr: true},
		{name: "negative number", payload: `-36`, wantErr: true},
		{name: "fractional number", payload: `36.5`, wantErr: true},
		{name: "exponent number", payload: `36e0`, wantErr: true},
		{name: "malformed JSON", payload: `36x`, wantErr: true},
		{name: "trailing JSON", payload: `36 37`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got NumericString
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s, got NumericString %q", tt.payload, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("String() = %q, want %q", got.String(), tt.want)
			}
			if string(got) != tt.want {
				t.Errorf("NumericString = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSizeReadResponseNumericFields(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "numeric fields",
			payload: `{"response":[{"id":105,"name":"small","type":"shared","cores":1,"ram":2048,"hdd":50}]}`,
		},
		{
			name:    "string fields",
			payload: `{"response":[{"id":"105","name":"small","type":"shared","cores":"1","ram":"2048","hdd":"50"}]}`,
		},
		{
			name:    "mixed wire representation",
			payload: `{"response":[{"id":105,"name":"small","type":"shared","cores":"1","ram":2048,"hdd":"50"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SizesListResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Response) != 1 {
				t.Fatalf("got %d sizes, want 1", len(got.Response))
			}
			size := got.Response[0]
			if size.Id.String() != "105" {
				t.Errorf("Id.String() = %q, want %q", size.Id.String(), "105")
			}
			if size.Cores.String() != "1" {
				t.Errorf("Cores.String() = %q, want %q", size.Cores.String(), "1")
			}
			if size.RAM.String() != "2048" {
				t.Errorf("RAM.String() = %q, want %q", size.RAM.String(), "2048")
			}
			if size.Disk.String() != "50" {
				t.Errorf("Disk.String() = %q, want %q", size.Disk.String(), "50")
			}
		})
	}
}
