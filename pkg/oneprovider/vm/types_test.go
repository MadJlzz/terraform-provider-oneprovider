package vm

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestStringOrNumber(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{name: "string", payload: `"36"`, want: "36"},
		{name: "arbitrary string", payload: `"cores"`, want: "cores"},
		{name: "empty string", payload: `""`, want: ""},
		{name: "number", payload: `36`, want: "36"},
		{name: "fractional number", payload: `36.5`, want: "36.5"},
		{name: "exponent number", payload: `36e0`, want: "36e0"},
		{name: "exact large number", payload: `9007199254740993`, want: "9007199254740993"},
		{name: "null", payload: `null`, want: ""},
		{name: "boolean", payload: `true`, wantErr: true},
		{name: "object", payload: `{}`, wantErr: true},
		{name: "array", payload: `[]`, wantErr: true},
		{name: "malformed JSON", payload: `36x`, wantErr: true},
		{name: "trailing JSON", payload: `36 37`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got StringOrNumber
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s, got %q", tt.payload, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("String() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestAPIIDWireSeam(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{name: "number", payload: `36`, want: "36"},
		{name: "string", payload: `"36"`, want: "36"},
		{name: "opaque string", payload: `"vm-36"`, want: "vm-36"},
		{name: "empty string", payload: `""`, want: ""},
		{name: "null", payload: `null`, want: ""},
		{name: "negative number remains wire text", payload: `-36`, want: "-36"},
		{name: "fractional number remains wire text", payload: `36.5`, want: "36.5"},
		{name: "exponent number remains wire text", payload: `36e0`, want: "36e0"},
		{name: "exact large number", payload: `9007199254740993`, want: "9007199254740993"},
		{name: "boolean", payload: `true`, wantErr: true},
		{name: "object", payload: `{}`, wantErr: true},
		{name: "array", payload: `[]`, wantErr: true},
		{name: "malformed JSON", payload: `36x`, wantErr: true},
		{name: "trailing JSON", payload: `36 37`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got APIID
			err := json.Unmarshal([]byte(tt.payload), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s, got %q", tt.payload, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("String() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestResponseIDsAcceptStringNumberAndNull(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		getID   func() string
	}{
		{
			name:    "location",
			payload: `{"response":{"region":[{"id":null}]}}`,
			getID: func() string {
				var response LocationsListResponse
				if err := json.Unmarshal([]byte(`{"response":{"region":[{"id":null}]}}`), &response); err != nil {
					t.Fatalf("unmarshal location: %v", err)
				}
				return response.Response["region"][0].Id.String()
			},
		},
		{
			name:    "size",
			payload: `{"response":[{"id":null}]}`,
			getID: func() string {
				var response SizesListResponse
				if err := json.Unmarshal([]byte(`{"response":[{"id":null}]}`), &response); err != nil {
					t.Fatalf("unmarshal size: %v", err)
				}
				return response.Response[0].Id.String()
			},
		},
		{
			name:    "template",
			payload: `{"response":[{"id":null}]}`,
			getID: func() string {
				var response TemplatesListResponse
				if err := json.Unmarshal([]byte(`{"response":[{"id":null}]}`), &response); err != nil {
					t.Fatalf("unmarshal template: %v", err)
				}
				return response.Templates[0].Id.String()
			},
		},
		{
			name:    "create",
			payload: `{"response":{"id":null}}`,
			getID: func() string {
				var response InstanceCreateResponse
				if err := json.Unmarshal([]byte(`{"response":{"id":null}}`), &response); err != nil {
					t.Fatalf("unmarshal create: %v", err)
				}
				return response.Response.Id.String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.getID(); got != "" {
				t.Errorf("ID = %q, want empty wire value", got)
			}
		})
	}
}

func TestLocationAvailableSizesWireRepresentations(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    []string
	}{
		{name: "strings", payload: `{"id":"1","available_sizes":["71","72"]}`, want: []string{"71", "72"}},
		{name: "numbers", payload: `{"id":1,"available_sizes":[71,72]}`, want: []string{"71", "72"}},
		{name: "mixed with null", payload: `{"id":1,"available_sizes":["71",72,null]}`, want: []string{"71", "72", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got LocationReadResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.AvailableSizes) != len(tt.want) {
				t.Fatalf("got %d sizes, want %d", len(got.AvailableSizes), len(tt.want))
			}
			for i, want := range tt.want {
				if got.AvailableSizes[i].String() != want {
					t.Errorf("AvailableSizes[%d] = %q, want %q", i, got.AvailableSizes[i], want)
				}
			}
		})
	}

	var nullResponse LocationReadResponse
	if err := json.Unmarshal([]byte(`{"available_sizes":null}`), &nullResponse); err != nil {
		t.Fatalf("unmarshal null available_sizes: %v", err)
	}
	if nullResponse.AvailableSizes != nil {
		t.Fatalf("null available_sizes became non-nil slice: %#v", nullResponse.AvailableSizes)
	}
}

func TestTemplateSizeAndDisplayOcaWireRepresentations(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantSize string
		wantOca  string
	}{
		{name: "string", payload: `{"id":"tpl","size":"5368709120","display":{"oca":"0"}}`, wantSize: "5368709120", wantOca: "0"},
		{name: "number", payload: `{"id":1,"size":5368709120,"display":{"oca":0}}`, wantSize: "5368709120", wantOca: "0"},
		{name: "null", payload: `{"id":null,"size":null,"display":{"oca":null}}`, wantSize: "", wantOca: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TemplateReadResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Size.String() != tt.wantSize {
				t.Errorf("Size = %q, want %q", got.Size, tt.wantSize)
			}
			if got.Display.Oca.String() != tt.wantOca {
				t.Errorf("Display.Oca = %q, want %q", got.Display.Oca, tt.wantOca)
			}
		})
	}
}

func TestInstanceReadResponseServerInstall(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    bool
	}{
		{name: "documented string", payload: `{"response":{"server_install":"false"}}`, want: false},
		{name: "boolean false", payload: `{"response":{"server_install":false}}`, want: false},
		{name: "boolean true", payload: `{"response":{"server_install":true}}`, want: true},
		{name: "string true", payload: `{"response":{"server_install":"true"}}`, want: true},
		{name: "null", payload: `{"response":{"server_install":null}}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got InstanceReadResponse
			if err := json.Unmarshal([]byte(tt.payload), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Response.ServerInstall != BoolOrString(tt.want) {
				t.Errorf("ServerInstall = %v, want %v", got.Response.ServerInstall, tt.want)
			}
		})
	}
}

func TestBoolOrStringRejectsUnsupportedValues(t *testing.T) {
	for _, payload := range []string{`"yes"`, `"null"`, `1`, `{}`, `[]`} {
		var got BoolOrString
		if err := json.Unmarshal([]byte(payload), &got); err == nil {
			t.Errorf("expected error for %s, got %v", payload, got)
		}
	}
}

func TestBoolOrStringAcceptsEscapedString(t *testing.T) {
	var got BoolOrString
	if err := json.Unmarshal([]byte(`"true"`), &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("BoolOrString = false, want true")
	}
}

// encoding/json treats null as a no-op for plain string and bool fields, and
// documents the same convention for Unmarshalers.
func TestNullLeavesValueUnchanged(t *testing.T) {
	got := struct {
		Number StringOrNumber
		ID     APIID
		Flag   BoolOrString
	}{Number: "1", ID: "2", Flag: true}

	if err := json.Unmarshal([]byte(`{"Number":null,"ID":null,"Flag":null}`), &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Number != "1" || got.ID != "2" || !got.Flag {
		t.Errorf("null changed values: %+v", got)
	}
}

func TestUnsupportedValueErrorDoesNotEchoPayload(t *testing.T) {
	var got APIID
	err := json.Unmarshal([]byte(`{"secret":"`+strings.Repeat("a", 500)+`"}`), &got)
	if err == nil {
		t.Fatal("expected error for object ID")
	}
	if strings.Contains(err.Error(), "secret") || len(err.Error()) > 100 {
		t.Errorf("error echoes payload: %q", err)
	}
}

// UnmarshalJSON can be called directly, bypassing the validation that
// json.Unmarshal performs before invoking it.
func TestDirectUnmarshalJSONRejectsInvalidInput(t *testing.T) {
	for _, payload := range []string{``, `36x`, `"36`, `nul`, `-`, `tru`, " 36", "36 "} {
		var number StringOrNumber
		if err := number.UnmarshalJSON([]byte(payload)); err == nil {
			t.Errorf("StringOrNumber.UnmarshalJSON(%q) succeeded with %q", payload, number)
		}
		var id APIID
		if err := id.UnmarshalJSON([]byte(payload)); err == nil {
			t.Errorf("APIID.UnmarshalJSON(%q) succeeded with %q", payload, id)
		}
		var flag BoolOrString
		if err := flag.UnmarshalJSON([]byte(payload)); err == nil {
			t.Errorf("BoolOrString.UnmarshalJSON(%q) succeeded with %v", payload, flag)
		}
	}
}

func TestInstanceCreateRequestUrlValuesKeepsIntContract(t *testing.T) {
	request := InstanceCreateRequest{
		LocationId:     33,
		InstanceSizeId: 45,
		TemplateId:     "image-uuid",
		Hostname:       "vm.example",
		SshKeys:        []string{"key-a", "key-b"},
	}
	want := url.Values{
		"location_id":   []string{"33"},
		"instance_size": []string{"45"},
		"template":      []string{"image-uuid"},
		"hostname":      []string{"vm.example"},
		"ssh_keys[0]":   []string{"key-a"},
		"ssh_keys[1]":   []string{"key-b"},
	}
	if got := request.UrlValues(); !reflect.DeepEqual(got, want) {
		t.Fatalf("UrlValues() = %#v, want %#v", got, want)
	}
}
