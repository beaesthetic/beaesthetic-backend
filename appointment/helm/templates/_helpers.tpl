{{- define "appointment.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "appointment.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $env := include "appointment.environment" . -}}
{{- if $env -}}
{{- printf "%s-%s" .Values.namespace $base -}}
{{- else -}}
{{- $base -}}
{{- end -}}
{{- end -}}

{{- define "appointment.renderEnvConfig" -}}
{{- $environment := include "appointment.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}

{{- define "appointment.commonLabels" -}}
{{- tpl (toYaml .Values.labels) . }}
{{- $environment := include "appointment.environment" . -}}
{{- if $environment }}
environment: {{ $environment | quote }}
{{- end }}
{{- end -}}

{{- define "appointment.rabbitmqVhost" -}}
{{- $environment := default "prod" (include "appointment.environment" .) -}}
{{- $namespaceBase := trimSuffix (printf "-%s" $environment) .Values.namespace -}}
{{- printf "%s-%s" $namespaceBase $environment -}}
{{- end -}}


