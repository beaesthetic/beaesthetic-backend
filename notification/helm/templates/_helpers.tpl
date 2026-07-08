{{- define "notification.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "notification.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $environment := default "prod" (include "notification.environment" .) -}}
{{- $namespaceBase := trimSuffix (printf "-%s" $environment) .Values.namespace -}}
{{- printf "%s-%s-%s" $namespaceBase $environment $base -}}
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

{{- define "notification.rabbitmqVhost" -}}
{{- $environment := default "prod" (include "notification.environment" .) -}}
{{- $namespaceBase := trimSuffix (printf "-%s" $environment) .Values.namespace -}}
{{- printf "%s-%s" $namespaceBase $environment -}}
{{- end -}}

