package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

// ---- unraid_vars ----

type varsModel struct {
	ID                 types.String `tfsdk:"id"`
	Version            types.String `tfsdk:"version"`
	Name               types.String `tfsdk:"name"`
	Comment            types.String `tfsdk:"comment"`
	TimeZone           types.String `tfsdk:"time_zone"`
	Workgroup          types.String `tfsdk:"workgroup"`
	Domain             types.String `tfsdk:"domain"`
	LocalTld           types.String `tfsdk:"local_tld"`
	UseSsl             types.Bool   `tfsdk:"use_ssl"`
	Port               types.Int64  `tfsdk:"port"`
	PortSsl            types.Int64  `tfsdk:"port_ssl"`
	UseSsh             types.Bool   `tfsdk:"use_ssh"`
	PortSsh            types.Int64  `tfsdk:"port_ssh"`
	UseTelnet          types.Bool   `tfsdk:"use_telnet"`
	UseNtp             types.Bool   `tfsdk:"use_ntp"`
	NtpServers         []string     `tfsdk:"ntp_servers"`
	StartArray         types.Bool   `tfsdk:"start_array"`
	SpindownDelay      types.String `tfsdk:"spindown_delay"`
	DefaultFsType      types.String `tfsdk:"default_fs_type"`
	ShareSmbEnabled    types.Bool   `tfsdk:"share_smb_enabled"`
	ShareNfsEnabled    types.Bool   `tfsdk:"share_nfs_enabled"`
	ShareCacheEnabled  types.Bool   `tfsdk:"share_cache_enabled"`
	ShareMoverSchedule types.String `tfsdk:"share_mover_schedule"`
	ShareCount         types.Int64  `tfsdk:"share_count"`
	MdState            types.String `tfsdk:"md_state"`
	FsState            types.String `tfsdk:"fs_state"`
	RegistrationType   types.String `tfsdk:"registration_type"`
}

func newVarsDataSource() datasource.DataSource {
	return &readOnlyDataSource[varsModel]{
		name: "vars",
		schema: schema.Schema{
			MarkdownDescription: "Selected system variables (`var.ini`) of the Unraid server. Requires `READ_ANY` on `VARS`.",
			Attributes: map[string]schema.Attribute{
				"id":                   cStr("Identifier of the vars object."),
				"version":              cStr("Unraid OS version."),
				"name":                 cStr("Server hostname."),
				"comment":              cStr("Server description."),
				"time_zone":            cStr("Configured time zone."),
				"workgroup":            cStr("SMB workgroup."),
				"domain":               cStr("Domain name."),
				"local_tld":            cStr("Local top-level domain."),
				"use_ssl":              cBool("Whether the web UI uses SSL."),
				"port":                 cInt("Web UI HTTP port."),
				"port_ssl":             cInt("Web UI HTTPS port."),
				"use_ssh":              cBool("Whether SSH is enabled."),
				"port_ssh":             cInt("SSH port."),
				"use_telnet":           cBool("Whether telnet is enabled."),
				"use_ntp":              cBool("Whether NTP time sync is enabled."),
				"ntp_servers":          cStrList("Configured NTP servers 1-4 (empty strings for unused slots)."),
				"start_array":          cBool("Whether the array starts automatically on boot."),
				"spindown_delay":       cStr("Default disk spin-down delay."),
				"default_fs_type":      cStr("Default file system for new disks."),
				"share_smb_enabled":    cBool("Whether SMB is enabled."),
				"share_nfs_enabled":    cBool("Whether NFS is enabled."),
				"share_cache_enabled":  cBool("Whether cache is enabled for shares."),
				"share_mover_schedule": cStr("Mover schedule (cron expression)."),
				"share_count":          cInt("Number of user shares."),
				"md_state":             cStr("Array (md) state."),
				"fs_state":             cStr("File system state."),
				"registration_type":    cStr("License type."),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (varsModel, error) {
			r, err := client.GetVars(ctx, c)
			if err != nil {
				return varsModel{}, err
			}
			v := r.Vars
			ntp := []string{}
			for _, s := range []*string{v.NtpServer1, v.NtpServer2, v.NtpServer3, v.NtpServer4} {
				ntp = append(ntp, types.StringPointerValue(s).ValueString())
			}
			return varsModel{
				ID:                 types.StringValue(v.Id),
				Version:            types.StringPointerValue(v.Version),
				Name:               types.StringPointerValue(v.Name),
				Comment:            types.StringPointerValue(v.Comment),
				TimeZone:           types.StringPointerValue(v.TimeZone),
				Workgroup:          types.StringPointerValue(v.Workgroup),
				Domain:             types.StringPointerValue(v.Domain),
				LocalTld:           types.StringPointerValue(v.LocalTld),
				UseSsl:             types.BoolPointerValue(v.UseSsl),
				Port:               intPtr(v.Port),
				PortSsl:            intPtr(v.Portssl),
				UseSsh:             types.BoolPointerValue(v.UseSsh),
				PortSsh:            intPtr(v.Portssh),
				UseTelnet:          types.BoolPointerValue(v.UseTelnet),
				UseNtp:             types.BoolPointerValue(v.UseNtp),
				NtpServers:         ntp,
				StartArray:         types.BoolPointerValue(v.StartArray),
				SpindownDelay:      types.StringPointerValue(v.SpindownDelay),
				DefaultFsType:      types.StringPointerValue(v.DefaultFsType),
				ShareSmbEnabled:    types.BoolPointerValue(v.ShareSmbEnabled),
				ShareNfsEnabled:    types.BoolPointerValue(v.ShareNfsEnabled),
				ShareCacheEnabled:  types.BoolPointerValue(v.ShareCacheEnabled),
				ShareMoverSchedule: types.StringPointerValue(v.ShareMoverSchedule),
				ShareCount:         intPtr(v.ShareCount),
				MdState:            types.StringPointerValue(v.MdState),
				FsState:            types.StringPointerValue(v.FsState),
				RegistrationType:   strEnum(v.RegTy),
			}, nil
		},
	}
}

// ---- unraid_server ----

type serverModel struct {
	ID       types.String `tfsdk:"id"`
	GUID     types.String `tfsdk:"guid"`
	Name     types.String `tfsdk:"name"`
	Comment  types.String `tfsdk:"comment"`
	Status   types.String `tfsdk:"status"`
	LanIP    types.String `tfsdk:"lan_ip"`
	LocalURL types.String `tfsdk:"local_url"`
}

func newServerDataSource() datasource.DataSource {
	return &readOnlyDataSource[serverModel]{
		name: "server",
		schema: schema.Schema{
			MarkdownDescription: "Identity of the Unraid server. Requires `READ_ANY` on `SERVERS`.",
			Attributes: map[string]schema.Attribute{
				"id":        cStr("Server identifier."),
				"guid":      cStr("Flash drive GUID."),
				"name":      cStr("Server name."),
				"comment":   cStr("Server description."),
				"status":    cStr("Connect status: `ONLINE`, `OFFLINE` or `NEVER_CONNECTED`."),
				"lan_ip":    cStr("LAN IP address."),
				"local_url": cStr("Local URL of the web UI."),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (serverModel, error) {
			r, err := client.GetServer(ctx, c)
			if err != nil {
				return serverModel{}, err
			}
			if r.Server == nil {
				return serverModel{}, errNotFound("server")
			}
			s := r.Server
			return serverModel{
				ID:       types.StringValue(s.Id),
				GUID:     types.StringValue(s.Guid),
				Name:     types.StringValue(s.Name),
				Comment:  types.StringPointerValue(s.Comment),
				Status:   types.StringValue(string(s.Status)),
				LanIP:    types.StringValue(s.Lanip),
				LocalURL: types.StringValue(s.Localurl),
			}, nil
		},
	}
}

// ---- unraid_system_time ----

type systemTimeModel struct {
	CurrentTime types.String `tfsdk:"current_time"`
	TimeZone    types.String `tfsdk:"time_zone"`
	UseNtp      types.Bool   `tfsdk:"use_ntp"`
	NtpServers  []string     `tfsdk:"ntp_servers"`
}

func newSystemTimeDataSource() datasource.DataSource {
	return &readOnlyDataSource[systemTimeModel]{
		name: "system_time",
		schema: schema.Schema{
			MarkdownDescription: "System time configuration. Requires `READ_ANY` on `VARS`.",
			Attributes: map[string]schema.Attribute{
				"current_time": cStr("Current server time (ISO-8601, UTC)."),
				"time_zone":    cStr("IANA time zone in use."),
				"use_ntp":      cBool("Whether NTP/PTP sync is enabled."),
				"ntp_servers":  cStrList("Configured NTP servers (empty strings for unused slots)."),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (systemTimeModel, error) {
			r, err := client.GetSystemTime(ctx, c)
			if err != nil {
				return systemTimeModel{}, err
			}
			t := r.SystemTime
			return systemTimeModel{
				CurrentTime: types.StringValue(t.CurrentTime),
				TimeZone:    types.StringValue(t.TimeZone),
				UseNtp:      types.BoolValue(t.UseNtp),
				NtpServers:  t.NtpServers,
			}, nil
		},
	}
}

// ---- unraid_ups_configuration ----

type upsConfigurationModel struct {
	Service             types.String `tfsdk:"service"`
	UpsCable            types.String `tfsdk:"ups_cable"`
	CustomUpsCable      types.String `tfsdk:"custom_ups_cable"`
	UpsType             types.String `tfsdk:"ups_type"`
	Device              types.String `tfsdk:"device"`
	OverrideUpsCapacity types.Int64  `tfsdk:"override_ups_capacity"`
	BatteryLevel        types.Int64  `tfsdk:"battery_level"`
	Minutes             types.Int64  `tfsdk:"minutes"`
	Timeout             types.Int64  `tfsdk:"timeout"`
	KillUps             types.String `tfsdk:"kill_ups"`
	NisIP               types.String `tfsdk:"nis_ip"`
	NetServer           types.String `tfsdk:"net_server"`
	UpsName             types.String `tfsdk:"ups_name"`
	ModelName           types.String `tfsdk:"model_name"`
}

func newUpsConfigurationDataSource() datasource.DataSource {
	return &readOnlyDataSource[upsConfigurationModel]{
		name: "ups_configuration",
		schema: schema.Schema{
			MarkdownDescription: "UPS (apcupsd) configuration. Values are returned as stored, typically lower-case.",
			Attributes: map[string]schema.Attribute{
				"service":               cStr("`enable` or `disable`."),
				"ups_cable":             cStr("Cable type, e.g. `usb`, `smart`, `ether`, `custom`."),
				"custom_ups_cable":      cStr("Custom cable string (when `ups_cable` is `custom`)."),
				"ups_type":              cStr("Communication type, e.g. `usb`, `net`, `snmp`."),
				"device":                cStr("Device path or network address."),
				"override_ups_capacity": cInt("Overridden UPS capacity."),
				"battery_level":         cInt("Battery percentage that triggers shutdown."),
				"minutes":               cInt("Remaining runtime (minutes) that triggers shutdown."),
				"timeout":               cInt("Seconds on battery before shutdown; 0 disables."),
				"kill_ups":              cStr("`yes` or `no`: power off the UPS after shutdown."),
				"nis_ip":                cStr("Network information server listen address."),
				"net_server":            cStr("`on` or `off`: allow network clients."),
				"ups_name":              cStr("UPS name for network monitoring."),
				"model_name":            cStr("Overridden UPS model name."),
			},
		},
		read: func(ctx context.Context, c graphql.Client) (upsConfigurationModel, error) {
			r, err := client.GetUpsConfiguration(ctx, c)
			if err != nil {
				return upsConfigurationModel{}, err
			}
			u := r.UpsConfiguration
			return upsConfigurationModel{
				Service:             types.StringPointerValue(u.Service),
				UpsCable:            types.StringPointerValue(u.UpsCable),
				CustomUpsCable:      types.StringPointerValue(u.CustomUpsCable),
				UpsType:             types.StringPointerValue(u.UpsType),
				Device:              types.StringPointerValue(u.Device),
				OverrideUpsCapacity: intPtr(u.OverrideUpsCapacity),
				BatteryLevel:        intPtr(u.BatteryLevel),
				Minutes:             intPtr(u.Minutes),
				Timeout:             intPtr(u.Timeout),
				KillUps:             types.StringPointerValue(u.KillUps),
				NisIP:               types.StringPointerValue(u.NisIp),
				NetServer:           types.StringPointerValue(u.NetServer),
				UpsName:             types.StringPointerValue(u.UpsName),
				ModelName:           types.StringPointerValue(u.ModelName),
			}, nil
		},
	}
}

// ---- unraid_settings ----

type settingsModel struct {
	ID         types.String `tfsdk:"id"`
	ValuesJSON types.String `tfsdk:"values_json"`
}

func newSettingsDataSource() datasource.DataSource {
	return &readOnlyDataSource[settingsModel]{
		name: "settings",
		schema: schema.Schema{
			MarkdownDescription: "Unified API settings (API origins, sandbox, SSO/OIDC, ...) as a JSON document. " +
				"Marked sensitive because it can contain OIDC client secrets. Use `jsondecode()` to access values.",
			Attributes: map[string]schema.Attribute{
				"id": cStr("Settings identifier."),
				"values_json": schema.StringAttribute{
					Computed:            true,
					Sensitive:           true,
					MarkdownDescription: "Current settings values as a JSON string.",
				},
			},
		},
		read: func(ctx context.Context, c graphql.Client) (settingsModel, error) {
			r, err := client.GetUnifiedSettings(ctx, c)
			if err != nil {
				return settingsModel{}, err
			}
			u := r.Settings.Unified
			return settingsModel{ID: types.StringValue(u.Id), ValuesJSON: types.StringValue(string(u.Values))}, nil
		},
	}
}
