{{- define "notification.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "notification.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $env := include "notification.environment" . -}}
{{- if $env -}}
{{- printf "%s_%s" $env $base -}}
{{- else -}}
{{- $base -}}
{{- end -}}
{{- end -}}

{{- define "notification.renderEnvConfig" -}}
{{- $environment := include "notification.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}
{{- define "notification.commonLabels" -}}
{{- tpl (toYaml .Values.labels) . }}
{{- $environment := include "notification.environment" . -}}
{{- if $environment }}
environment: {{ $environment | quote }}
{{- end }}
{{- end -}}


