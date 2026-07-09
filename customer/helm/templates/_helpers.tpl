{{- define "customer.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "customer.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $environment := default "prod" (include "customer.environment" .) -}}
{{- $namespaceBase := trimSuffix (printf "-%s" $environment) .Values.namespace -}}
{{- printf "%s-%s-%s" $namespaceBase $environment $base -}}
{{- end -}}

{{- define "customer.renderEnvConfig" -}}
{{- $environment := include "customer.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}

{{- define "customer.commonLabels" -}}
{{- tpl (toYaml .Values.labels) . }}
{{- $environment := include "customer.environment" . -}}
{{- if $environment }}
environment: {{ $environment | quote }}
{{- end }}
{{- end -}}