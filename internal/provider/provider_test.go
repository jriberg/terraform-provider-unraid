package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestResolveConfigFailsClosed(t *testing.T) {
	t.Setenv("UNRAID_ENDPOINT", "")
	t.Setenv("UNRAID_API_KEY", "")
	cases := map[string]struct {
		endpoint, key string
		wantErrors    int
	}{
		"missing both":    {"", "", 2},
		"missing key":     {"https://tower/graphql", "", 1},
		"relative url":    {"tower/graphql", "k", 1},
		"bad scheme":      {"ftp://tower/graphql", "k", 1},
		"valid https":     {"https://tower/graphql", "k", 0},
		"http only warns": {"http://tower/graphql", "k", 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, diags := resolveConfig(providerModel{Endpoint: types.StringValue(tc.endpoint), APIKey: types.StringValue(tc.key)})
			if got := diags.ErrorsCount(); got != tc.wantErrors {
				t.Fatalf("errors = %d, want %d: %v", got, tc.wantErrors, diags)
			}
		})
	}
}

func TestResolveConfigEnvFallback(t *testing.T) {
	t.Setenv("UNRAID_ENDPOINT", "https://env/graphql")
	t.Setenv("UNRAID_API_KEY", "envkey")
	e, k, diags := resolveConfig(providerModel{Endpoint: types.StringNull(), APIKey: types.StringNull()})
	if diags.HasError() || e != "https://env/graphql" || k != "envkey" {
		t.Fatalf("got %q %q %v", e, k, diags)
	}
}

// fakeUnraid answers GraphQL operations by name with canned data and rejects wrong API keys.
func fakeUnraid(t *testing.T, apiKey string, responses map[string]string) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != apiKey {
			http.Error(w, `{"errors":[{"message":"Unauthorized"}]}`, http.StatusUnauthorized)
			return
		}
		var req struct {
			OperationName string `json:"operationName"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		data, ok := responses[req.OperationName]
		if !ok {
			t.Errorf("unexpected operation %q", req.OperationName)
			http.Error(w, "unknown op", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"data":` + data + `}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

var factories = map[string]func() (tfprotov6.ProviderServer, error){
	"unraid": providerserver.NewProtocol6WithError(New("test")()),
}

func providerBlock(endpoint, key string) string {
	return `provider "unraid" {
  endpoint = "` + endpoint + `"
  api_key  = "` + key + `"
}
`
}

func TestDataSourcesAgainstFakeServer(t *testing.T) {
	srv := fakeUnraid(t, "secret", map[string]string{
		"GetShares": `{"shares":[{"id":"s:appdata","name":"appdata","comment":null,"free":"1000","used":20,"size":null,
			"include":["disk1"],"exclude":null,"cache":true,"allocator":"highwater","splitLevel":"","floor":"0","cow":"auto","luksStatus":"0"}]}`,
		"GetArray": `{"array":{"id":"a","state":"STARTED","capacity":{"kilobytes":{"free":"1","used":"2","total":"3"}},
			"parities":[{"id":"p","idx":0,"name":"parity","device":"sdb","size":100,"status":"DISK_OK","type":"PARITY","rotational":true,"temp":null,
			"fsType":null,"fsSize":null,"fsFree":null,"fsUsed":null,"numErrors":0,"isSpinning":false}],"disks":[],"caches":[],"bootDevices":[]}}`,
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config: providerBlock(srv.URL, "secret") + `
data "unraid_shares" "all" {}
data "unraid_array" "this" {}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.unraid_shares.all", "shares.#", "1"),
				resource.TestCheckResourceAttr("data.unraid_shares.all", "shares.0.name", "appdata"),
				resource.TestCheckResourceAttr("data.unraid_shares.all", "shares.0.free_kb", "1000"),
				resource.TestCheckResourceAttr("data.unraid_shares.all", "shares.0.used_kb", "20"),
				resource.TestCheckNoResourceAttr("data.unraid_shares.all", "shares.0.size_kb"),
				resource.TestCheckResourceAttr("data.unraid_shares.all", "shares.0.include.0", "disk1"),
				resource.TestCheckResourceAttr("data.unraid_array.this", "state", "STARTED"),
				resource.TestCheckResourceAttr("data.unraid_array.this", "parities.0.status", "DISK_OK"),
				resource.TestCheckResourceAttr("data.unraid_array.this", "disks.#", "0"),
			),
		}},
	})
}

func TestForbiddenGivesPermissionHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Forbidden resource","path":["apiKeys"]}]}`))
	}))
	t.Cleanup(srv.Close)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config:      providerBlock(srv.URL, "k") + `data "unraid_api_keys" "all" {}`,
			ExpectError: regexp.MustCompile(`lacks the permission`),
		}},
	})
}

func TestWrongAPIKeyFails(t *testing.T) {
	srv := fakeUnraid(t, "secret", nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{{
			Config:      providerBlock(srv.URL, "wrong") + `data "unraid_vars" "this" {}`,
			ExpectError: regexp.MustCompile(`(?i)unauthorized|401`),
		}},
	})
}

// Every schema attribute must carry a description so generated registry docs are complete.
func TestAllDataSourceAttributesDocumented(t *testing.T) {
	p := New("test")()
	for _, f := range p.DataSources(t.Context()) {
		ds := f().(interface{ schemaForTest() map[string]string })
		for attr, desc := range ds.schemaForTest() {
			if strings.TrimSpace(desc) == "" {
				t.Errorf("attribute %s has no description", attr)
			}
		}
	}
}
