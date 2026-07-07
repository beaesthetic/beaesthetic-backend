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

{{- define "notification.rabbitmqVhost" -}}
{{- $env := include "notification.environment" . -}}
{{- if $env -}}
{{- printf "beaesthetic-%s" $env -}}
{{- end -}}
{{- end -}}
{{- define "notification.renderEnvConfig" -}}
{{- $environment := include "notification.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}

