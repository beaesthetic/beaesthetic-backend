{{- define "scheduler.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "scheduler.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $env := include "scheduler.environment" . -}}
{{- if $env -}}
{{- printf "%s-%s" .Values.namespace $base -}}
{{- else -}}
{{- $base -}}
{{- end -}}
{{- end -}}

{{- define "scheduler.renderEnvConfig" -}}
{{- $environment := include "scheduler.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}

{{- define "scheduler.commonLabels" -}}
{{- tpl (toYaml .Values.labels) . }}
{{- $environment := include "scheduler.environment" . -}}
{{- if $environment }}
environment: {{ $environment | quote }}
{{- end }}
{{- end -}}

{{- define "scheduler.rabbitmqVhost" -}}
{{- $environment := default "prod" (include "scheduler.environment" .) -}}
{{- $namespaceBase := trimSuffix (printf "-%s" $environment) .Values.namespace -}}
{{- printf "%s-%s" $namespaceBase $environment -}}
{{- end -}}


