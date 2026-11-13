package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

// readOnlyDataSource is a data source that takes no arguments and maps one GraphQL query to a model.
type readOnlyDataSource[M any] struct {
	name   string
	schema schema.Schema
	read   func(context.Context, graphql.Client) (M, error)
	client graphql.Client
}

func (d *readOnlyDataSource[M]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *readOnlyDataSource[M]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema
}

func (d *readOnlyDataSource[M]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(graphql.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected a graphql.Client; this is a provider bug.")
		return
	}
	d.client = c
}

func (d *readOnlyDataSource[M]) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	m, err := d.read(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read unraid_"+d.name, withPermissionHint(err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// Conversion helpers from generated nullable GraphQL fields to framework values.

func strEnum[T ~string](p *T) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*p))
}

func intPtr(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}

func bigPtr(p *client.BigInt) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}

func enumList[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return out
}

// Schema shorthands; every attribute is computed in read-only data sources.

func cStr(desc string) schema.StringAttribute {
	return schema.StringAttribute{Computed: true, MarkdownDescription: desc}
}

func cInt(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{Computed: true, MarkdownDescription: desc}
}

func cBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{Computed: true, MarkdownDescription: desc}
}

func cFloat(desc string) schema.Float64Attribute {
	return schema.Float64Attribute{Computed: true, MarkdownDescription: desc}
}

func cStrList(desc string) schema.ListAttribute {
	return schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: desc}
}

func cObjList(desc string, attrs map[string]schema.Attribute) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{Computed: true, MarkdownDescription: desc, NestedObject: schema.NestedAttributeObject{Attributes: attrs}}
}

// schemaForTest flattens attribute descriptions (including nested ones) for the documentation test.
func (d *readOnlyDataSource[M]) schemaForTest() map[string]string {
	out := map[string]string{d.name: d.schema.MarkdownDescription}
	var walk func(prefix string, attrs map[string]schema.Attribute)
	walk = func(prefix string, attrs map[string]schema.Attribute) {
		for k, a := range attrs {
			out[prefix+k] = a.GetMarkdownDescription()
			if n, ok := a.(schema.ListNestedAttribute); ok {
				walk(prefix+k+".", n.NestedObject.Attributes)
			}
		}
	}
	walk(d.name+".", d.schema.Attributes)
	return out
}
