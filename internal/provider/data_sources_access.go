package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

// ---- unraid_api_keys ----

type permissionModel struct {
	Resource types.String `tfsdk:"resource"`
	Actions  []string     `tfsdk:"actions"`
}

type apiKeyModel struct {
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	Roles       []string          `tfsdk:"roles"`
	CreatedAt   types.String      `tfsdk:"created_at"`
	Permissions []permissionModel `tfsdk:"permissions"`
}

type apiKeysModel struct {
	APIKeys []apiKeyModel `tfsdk:"api_keys"`
}

func newAPIKeysDataSource() datasource.DataSource {
	return &readOnlyDataSource[apiKeysModel]{
		name: "api_keys",
		schema: schema.Schema{
			MarkdownDescription: "Metadata of all API keys. The secret key values are never read. Requires `READ_ANY` on `API_KEY`.",
			Attributes: map[string]schema.Attribute{
				"api_keys": cObjList("List of API keys.", map[string]schema.Attribute{
					"id":          cStr("API key identifier."),
					"name":        cStr("API key name."),
					"description": cStr("API key description."),
					"roles":       cStrList("Roles: `ADMIN`, `CONNECT`, `GUEST`, `VIEWER`."),
					"created_at":  cStr("Creation timestamp."),
					"permissions": cObjList("Explicit permissions.", map[string]schema.Attribute{
						"resource": cStr("Resource, e.g. `DOCKER`."),
						"actions":  cStrList("Actions, e.g. `READ_ANY`."),
					}),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (apiKeysModel, error) {
			r, err := client.GetApiKeys(ctx, c)
			if err != nil {
				return apiKeysModel{}, err
			}
			out := apiKeysModel{APIKeys: []apiKeyModel{}}
			for _, k := range r.ApiKeys {
				perms := []permissionModel{}
				for _, p := range k.Permissions {
					perms = append(perms, permissionModel{Resource: types.StringValue(string(p.Resource)), Actions: enumList(p.Actions)})
				}
				out.APIKeys = append(out.APIKeys, apiKeyModel{
					ID:          types.StringValue(k.Id),
					Name:        types.StringValue(k.Name),
					Description: types.StringPointerValue(k.Description),
					Roles:       enumList(k.Roles),
					CreatedAt:   types.StringValue(k.CreatedAt),
					Permissions: perms,
				})
			}
			return out, nil
		},
	}
}

// ---- unraid_rclone_remotes ----

type rcloneRemoteModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type rcloneRemotesModel struct {
	Remotes []rcloneRemoteModel `tfsdk:"remotes"`
}

func newRCloneRemotesDataSource() datasource.DataSource {
	return &readOnlyDataSource[rcloneRemotesModel]{
		name: "rclone_remotes",
		schema: schema.Schema{
			MarkdownDescription: "Configured rclone remotes (name and type only; parameters contain credentials and are not read). Requires `READ_ANY` on `FLASH`.",
			Attributes: map[string]schema.Attribute{
				"remotes": cObjList("List of rclone remotes.", map[string]schema.Attribute{
					"name": cStr("Remote name."),
					"type": cStr("Backend type, e.g. `s3` or `drive`."),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (rcloneRemotesModel, error) {
			r, err := client.GetRCloneRemotes(ctx, c)
			if err != nil {
				return rcloneRemotesModel{}, err
			}
			out := rcloneRemotesModel{Remotes: []rcloneRemoteModel{}}
			for _, rm := range r.Rclone.Remotes {
				out.Remotes = append(out.Remotes, rcloneRemoteModel{Name: types.StringValue(rm.Name), Type: types.StringValue(rm.Type)})
			}
			return out, nil
		},
	}
}
