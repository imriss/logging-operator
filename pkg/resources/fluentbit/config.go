/*
 * Copyright © 2019 Banzai Cloud
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fluentbit

var fluentBitConfigTemplate = `
[SERVICE]
    Flush        1
    Daemon       Off
    Log_Level    info
    Parsers_File parsers.conf
    HTTP_Server  On
    HTTP_Listen  0.0.0.0
    HTTP_Port    {{ .Monitor.Port }}

[INPUT]
    Name             tail
    Path             /var/log/containers/*.log
    Parser           docker
    Tag              kubernetes.*
    Refresh_Interval 5
    Mem_Buf_Limit    5MB
    Skip_Long_Lines  On
    DB               /tail-db/tail-containers-state.db
    DB.Sync          Normal

[FILTER]
    Name                kubernetes
    Match               kubernetes.*
    Kube_Tag_Prefix     kubernetes.var.log.containers.
    Kube_URL            https://kubernetes.default.svc:443
    Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
    Merge_Log           On

[FILTER]
    Name    lua
    Match   kube.*
    script  /fluent-bit/etc/functions.lua
    call    dedot

[OUTPUT]
    Name          forward
    Match         *
    Host          fluentd.{{ .Namespace }}.svc
    Port          24240
    {{ if .TLS.Enabled }}
    tls           On
    tls.verify    Off
    tls.ca_file   /fluent-bit/tls/caCert
    tls.crt_file  /fluent-bit/tls/clientCert
    tls.key_file  /fluent-bit/tls/clientKey
    Shared_Key    {{ .TLS.SharedKey }}
    {{- end }}
    Retry_Limit   False
`
var fluentBitLuaFunctionsTemplate = `
function dedot(tag, timestamp, record)
    if record["kubernetes"] == nil then
        return 0, 0, 0
    end
    dedot_keys(record["kubernetes"]["annotations"])
    dedot_keys(record["kubernetes"]["labels"])
    return 1, timestamp, record
end

function dedot_keys(map)
    if map == nil then
        return
    end
    for k, v in pairs(map) do
        dedotted = string.gsub(k, "%.", "_")
        if k ~= dedotted then
            map[dedotted] = v
            map[k] = nil
        end
    end
end
`
