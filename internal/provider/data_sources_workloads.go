package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

// ---- unraid_docker_containers ----

type containerPortModel struct {
	IP          types.String `tfsdk:"ip"`
	PrivatePort types.Int64  `tfsdk:"private_port"`
	PublicPort  types.Int64  `tfsdk:"public_port"`
	Type        types.String `tfsdk:"type"`
}

type containerModel struct {
	ID                types.String         `tfsdk:"id"`
	Names             []string             `tfsdk:"names"`
	Image             types.String         `tfsdk:"image"`
	State             types.String         `tfsdk:"state"`
	Status            types.String         `tfsdk:"status"`
	AutoStart         types.Bool           `tfsdk:"auto_start"`
	AutoStartOrder    types.Int64          `tfsdk:"auto_start_order"`
	AutoStartWait     types.Int64          `tfsdk:"auto_start_wait"`
	NetworkMode       types.String         `tfsdk:"network_mode"`
	Ports             []containerPortModel `tfsdk:"ports"`
	IsUpdateAvailable types.Bool           `tfsdk:"is_update_available"`
	WebUIURL          types.String         `tfsdk:"web_ui_url"`
}

type containersModel struct {
	Containers []containerModel `tfsdk:"containers"`
}

func newDockerContainersDataSource() datasource.DataSource {
	return &readOnlyDataSource[containersModel]{
		name: "docker_containers",
		schema: schema.Schema{
			MarkdownDescription: "All Docker containers. Requires `READ_ANY` on `DOCKER`.",
			Attributes: map[string]schema.Attribute{
				"containers": cObjList("List of containers.", map[string]schema.Attribute{
					"id":               cStr("Container identifier."),
					"names":            cStrList("Container names (with leading `/`)."),
					"image":            cStr("Image reference."),
					"state":            cStr("`RUNNING`, `PAUSED` or `EXITED`."),
					"status":           cStr("Human-readable status."),
					"auto_start":       cBool("Whether the container auto-starts."),
					"auto_start_order": cInt("Zero-based position in the auto-start list."),
					"auto_start_wait":  cInt("Seconds to wait after starting."),
					"network_mode":     cStr("Docker network mode."),
					"ports": cObjList("Port mappings.", map[string]schema.Attribute{
						"ip":           cStr("Bound host IP."),
						"private_port": cInt("Container port."),
						"public_port":  cInt("Host port."),
						"type":         cStr("`TCP` or `UDP`."),
					}),
					"is_update_available": cBool("Whether a newer image is available."),
					"web_ui_url":          cStr("Resolved web UI URL from the template."),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (containersModel, error) {
			r, err := client.GetDockerContainers(ctx, c)
			if err != nil {
				return containersModel{}, err
			}
			out := containersModel{Containers: []containerModel{}}
			for _, ct := range r.Docker.Containers {
				ports := []containerPortModel{}
				for _, p := range ct.Ports {
					ports = append(ports, containerPortModel{
						IP:          types.StringPointerValue(p.Ip),
						PrivatePort: intPtr(p.PrivatePort),
						PublicPort:  intPtr(p.PublicPort),
						Type:        types.StringValue(string(p.Type)),
					})
				}
				network := types.StringNull()
				if ct.HostConfig != nil {
					network = types.StringValue(ct.HostConfig.NetworkMode)
				}
				out.Containers = append(out.Containers, containerModel{
					ID:                types.StringValue(ct.Id),
					Names:             ct.Names,
					Image:             types.StringValue(ct.Image),
					State:             types.StringValue(string(ct.State)),
					Status:            types.StringValue(ct.Status),
					AutoStart:         types.BoolValue(ct.AutoStart),
					AutoStartOrder:    intPtr(ct.AutoStartOrder),
					AutoStartWait:     intPtr(ct.AutoStartWait),
					NetworkMode:       network,
					Ports:             ports,
					IsUpdateAvailable: types.BoolPointerValue(ct.IsUpdateAvailable),
					WebUIURL:          types.StringPointerValue(ct.WebUiUrl),
				})
			}
			return out, nil
		},
	}
}

// ---- unraid_vms ----

type vmModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	State types.String `tfsdk:"state"`
}

type vmsModel struct {
	Vms []vmModel `tfsdk:"vms"`
}

func newVmsDataSource() datasource.DataSource {
	return &readOnlyDataSource[vmsModel]{
		name: "vms",
		schema: schema.Schema{
			MarkdownDescription: "All virtual machines. Requires `READ_ANY` on `VMS`.",
			Attributes: map[string]schema.Attribute{
				"vms": cObjList("List of VMs.", map[string]schema.Attribute{
					"id":    cStr("VM identifier (UUID)."),
					"name":  cStr("VM name."),
					"state": cStr("VM state, e.g. `RUNNING` or `SHUTOFF`."),
				}),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (vmsModel, error) {
			r, err := client.GetVms(ctx, c)
			if err != nil {
				return vmsModel{}, err
			}
			out := vmsModel{Vms: []vmModel{}}
			for _, v := range r.Vms.Domains {
				out.Vms = append(out.Vms, vmModel{
					ID:    types.StringValue(v.Id),
					Name:  types.StringPointerValue(v.Name),
					State: types.StringValue(string(v.State)),
				})
			}
			return out, nil
		},
	}
}
