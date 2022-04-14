//
//  a10_slb.go  --  SLB related aXAPI API calls
//
//  John D. Allen
//  Sr. Solutions Engineer
//  A10 Networks, Inc.
//
//  Copyright A10 Networks (c) 2020, All Rights Reserved.
//

package axapi

import (
	"strings"

	"github.com/tidwall/gjson"
)

//  GetSLBservers()
//-----------------------------------------------------------------------------
type Server struct {
	Name            string
	Host            string
	Status          string
	Template        string
	ConnectionLimit uint64
	Weight          int
}

// GetSLBservers()
//-----------------------------------------------------------------------------
func (d Device) GetSLBservers() ([]Server, error) {
	var s []Server
	body, err := _restCall(d, "/slb/server", "GET", nil)
	if err != nil {
		return s, err
	}
	if e, msg := d.chkResp(body); e {
		return s, msg
	}

	for _, v := range gjson.GetBytes(body, "server-list").Array() {
		var x Server
		x.Name = gjson.Get(v.String(), "name").Str
		x.Host = gjson.Get(v.String(), "host").Str
		x.Status = gjson.Get(v.String(), "action").Str
		x.Template = gjson.Get(v.String(), "template-server").Str
		x.ConnectionLimit = gjson.Get(v.String(), "conn-limit").Uint()
		x.Weight = int(gjson.Get(v.String(), "weight").Int())
		s = append(s, x)
	}

	return s, nil
}

// GetServceGroups()
//-----------------------------------------------------------------------------
type Member struct {
	Name     string
	Port     int
	State    string
	Priority int
}

type SvcGrp struct {
	Name        string
	Protocol    string
	LBMethod    string
	Healthcheck string
	Members     []Member
}

func (d Device) GetServiceGroups() ([]SvcGrp, error) {
	var sg []SvcGrp
	body, err := _restCall(d, "/slb/service-group-list", "GET", nil)
	if err != nil {
		return sg, err
	}
	if e, msg := d.chkResp(body); e {
		return sg, msg
	}

	for _, v := range gjson.GetBytes(body, "service-group-list").Array() {
		var x SvcGrp
		x.Name = gjson.Get(v.String(), "name").Str
		x.Protocol = gjson.Get(v.String(), "protocol").Str
		x.LBMethod = gjson.Get(v.String(), "lb-method").Str
		x.Healthcheck = gjson.Get(v.String(), "health-check").Str
		for _, z := range gjson.Get(v.String(), "member-list").Array() {
			var m Member
			m.Name = gjson.Get(z.String(), "name").Str
			m.Port = int(gjson.Get(z.String(), "port").Int())
			m.State = gjson.Get(z.String(), "member-state").Str
			m.Priority = int(gjson.Get(z.String(), "member-priority").Int())
			x.Members = append(x.Members, m)
		}
		sg = append(sg, x)
	}

	return sg, nil
}

// GetVSlist()
//-----------------------------------------------------------------------------
type Port struct {
	PortNumber int
	Protocol   string
	ConnLimit  uint64
	Status     string
	AutoSNAT   int
	SvcGrp     string
	Throughput uint64
}

type VS struct {
	Name   string
	IP     string
	Status string
	Ports  []Port
}

func (d Device) GetVSlist() ([]VS, error) {
	var vsl []VS
	body, err := _restCall(d, "/slb/virtual-server-list", "GET", nil)
	if err != nil {
		return vsl, err
	}
	if e, msg := d.chkResp(body); e {
		return vsl, msg
	}

	for _, s := range gjson.GetBytes(body, "virtual-server-list").Array() {
		var vs VS
		vs.Name = gjson.Get(s.String(), "name").Str
		vs.IP = gjson.Get(s.String(), "ip-address").Str
		vs.Status = gjson.Get(s.String(), "enable-disable-action").Str
		for _, v := range gjson.Get(s.String(), "port-list").Array() {
			var p Port
			p.PortNumber = int(gjson.Get(v.String(), "port-number").Int())
			p.Protocol = gjson.Get(v.String(), "protocol").Str
			p.ConnLimit = gjson.Get(v.String(), "conn-limit").Uint()
			p.Status = gjson.Get(v.String(), "action").Str
			p.AutoSNAT = int(gjson.Get(v.String(), "auto").Int())
			p.SvcGrp = gjson.Get(v.String(), "service-group").Str
			vs.Ports = append(vs.Ports, p)
		}
		vsl = append(vsl, vs)
	}

	return vsl, nil
}

// GetVSThroughput()
//-----------------------------------------------------------------------------
//     {
// 			"port": {
// 	"stats": {
// 		"curr_conn": 4,
// 		"total_l4_conn": 0,
// 		"total_l7_conn": 21,
// 		"total_tcp_conn": 21,
// 		"total_conn": 21,
// 		"total_fwd_bytes": 595742,
// 		"total_fwd_pkts": 9915,
// 		"total_rev_bytes": 3942725,
// 		"total_rev_pkts": 10326,
// 		"total_dns_pkts": 0,
// 		"total_mf_dns_pkts": 0,
// 		"es_total_failure_actions": 0,
// 		"compression_bytes_before": 0,
// 		"compression_bytes_after": 0,
// 		"compression_hit": 0,
// 		"compression_miss": 0,
// 		"compression_miss_no_client": 0,
// 		"compression_miss_template_exclusion": 0,
// 		"curr_req": 0,
// 		"total_req": 0,
// 		"total_req_succ": 0,
// 		"peak_conn": 0,
// 		"curr_conn_rate": 0,
// 		"last_rsp_time": 85,
// 		"fastest_rsp_time": 82,
// 		"slowest_rsp_time": 303,
// 		"loc_permit": 0,
// 		"loc_deny": 0,
// 		"loc_conn": 0,
// 		"curr_ssl_conn": 0,
// 		"total_ssl_conn": 0,
// 		"backend-time-to-first-byte": 0,
// 		"backend-time-to-last-byte": 0,
// 		"in-latency": 0,
// 		"out-latency": 0,
// 		"total_fwd_bytes_out": 491426,
// 		"total_fwd_pkts_out": 6226,
// 		"total_rev_bytes_out": 4091303,
// 		"total_rev_pkts_out": 10389,
// 		"curr_req_rate": 0,
// 		"curr_resp": 0,
// 		"total_resp": 0,
// 		"total_resp_succ": 0,
// 		"curr_resp_rate": 0,
// 		"curr_conn_overflow": 5,
// 		"dnsrrl_total_allowed": 0,
// 		"dnsrrl_total_dropped": 0,
// 		"dnsrrl_total_slipped": 0,
// 		"dnsrrl_bad_fqdn": 0,
// 		"throughput-bits-per-sec": 267184,
// 		"dynamic-memory": 0,
// 		"ip_only_lb_fwd_bytes": 0,
// 		"ip_only_lb_rev_bytes": 0,
// 		"ip_only_lb_fwd_pkts": 0,
// 		"ip_only_lb_rev_pkts": 0,
// 		"total_dns_filter_type_drop": 0,
// 		"total_dns_filter_class_drop": 0,
// 		"dns_filter_type_a_drop": 0,
// 		"dns_filter_type_aaaa_drop": 0,
// 		"dns_filter_type_cname_drop": 0,
// 		"dns_filter_type_mx_drop": 0,
// 		"dns_filter_type_ns_drop": 0,
// 		"dns_filter_type_srv_drop": 0,
// 		"dns_filter_type_ptr_drop": 0,
// 		"dns_filter_type_soa_drop": 0,
// 		"dns_filter_type_txt_drop": 0,
// 		"dns_filter_type_any_drop": 0,
// 		"dns_filter_type_others_drop": 0,
// 		"dns_filter_class_internet_drop": 0,
// 		"dns_filter_class_chaos_drop": 0,
// 		"dns_filter_class_hesiod_drop": 0,
// 		"dns_filter_class_none_drop": 0,
// 		"dns_filter_class_any_drop": 0,
// 		"dns_filter_class_others_drop": 0,
// 		"dns_rpz_action_drop": 0,
// 		"dns_rpz_action_pass_thru": 0,
// 		"dns_rpz_action_tcp_only": 0,
// 		"dns_rpz_action_nxdomain": 0,
// 		"dns_rpz_action_nodata": 0,
// 		"dns_rpz_action_local_data": 0,
// 		"dns_rpz_trigger_client_ip": 0,
// 		"dns_rpz_trigger_resp_ip": 0,
// 		"dns_rpz_trigger_ns_ip": 0,
// 		"dns_rpz_trigger_qname": 0,
// 		"dns_rpz_trigger_ns_name": 0
// 	},
// 	"a10-url": "/axapi/v3/slb/virtual-server/ws-vip/port/80+http/stats",
// 	"port-number": 80,
// 	"protocol": "http"
// }

func (d Device) GetVSThroughput(vs string, p string) (Port, error) {
	// p is in the format "80+http" to match the API call URL requirement.
	// Throughput returned is in bps
	var port Port
	url := "/slb/virtual-server/" + vs + "/port/" + p + "/stats"
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return port, err
	}
	if e, msg := d.chkResp(body); e {
		return port, msg
	}

	port.PortNumber = int(gjson.GetBytes(body, "port.port-number").Int())
	port.Protocol = gjson.GetBytes(body, "port.protocol").Str
	port.Throughput = gjson.GetBytes(body, "port.stats.throughput-bits-per-sec").Uint()

	return port, nil
}

// GetServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) GetServerTemplate(tpl string) (string, error) {
	url := "/slb/template/server/" + tpl
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return "", err
	}
	if e, msg := d.chkResp(body); e {
		return "", msg
	}
	return string(body), err
}

// GetVirtualServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) GetVirtualServerTemplate(tpl string) (string, error) {
	url := "/slb/template/virtual-server/" + tpl
	body, err := _restCall(d, url, "GET", nil)
	if err != nil {
		return "", err
	}
	if e, msg := d.chkResp(body); e {
		return "", msg
	}
	return string(body), err
}

// CreateServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) CreateServerTemplate(payload string) error {
	// Payload should have at least the 'name' field, and any attributes you want to set.
	// Example:
	// "server": {
	//     "name": "test",
	//     "conn-limit": 64000000,
	//     "conn-limit-no-logging": 0,
	//     "dns-query-interval": 10,
	//     "dns-fail-interval": 30,
	//     "dynamic-server-prefix": "DRS",
	//     "extended-stats": 0,
	//     "log-selection-failure": 0,
	//     "health-check-disable": 0,
	//     "max-dynamic-server": 255,
	//     "min-ttl-ratio": 2,
	//     "weight": 1,
	//     "spoofing-cache": 0,
	//     "stats-data-action": "stats-data-enable",
	//     "slow-start": 0,
	//     "bw-rate-limit-acct": "all",
	//     "bw-rate-limit": 1000,
	//     "bw-rate-limit-resume": 800,
	//     "bw-rate-limit-duration": 20,
	//     "bw-rate-limit-no-logging": 0
	// }
	url := "/slb/template/server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) UpdateServerTemplate(payload string) error {
	// NOTE: The 'name' field MUST be a part of the payload!
	url := "/slb/template/server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "PUT", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// CreateVirtualServerTemplate()
//-----------------------------------------------------------------------------
// Payload will at least need the 'name' field, and any KVs you want to set.
// Example:
// "virtual-server": {
// 	  "name": "test2",
// 	  "conn-limit": 200,
// 	  "conn-rate-limit": 200
// }
func (d Device) CreateVirtualServerTemplate(payload string) error {
	url := "/slb/template/virtual-server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateVirtualServerTemplate()
//-----------------------------------------------------------------------------
func (d Device) UpdateVirtualServerTemplate(payload string) error {
	// NOTE: The 'name' field MUST be a part of the payload!
	url := "/slb/template/virtual-server"
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "PUT", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}

// UpdateVirtualServer()
//-----------------------------------------------------------------------------
// This function only adds/updates to a virtual-server. It will overwrite vaules
// if they already exist, or add KV lines to the virtual-server config. It retains
// all other vaules (unlike a PUT would.)
func (d Device) UpdateVirtualServer(vs string, payload string) error {
	url := "/slb/virtual-server/" + vs
	pl := strings.NewReader(payload)
	body, err := _restCall(d, url, "POST", pl)
	if err != nil {
		return err
	}
	if e, msg := d.chkResp(body); e {
		return msg
	}
	return nil
}
