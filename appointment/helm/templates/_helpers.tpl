{{- define "appointment.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "appointment.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $env := include "appointment.environment" . -}}
{{- if $env -}}
{{- printf "%s_%s" $env $base -}}
{{- else -}}
{{- $base -}}
{{- end -}}
{{- end -}}

{{- define "appointment.rabbitmqVhost" -}}
{{- $env := include "appointment.environment" . -}}
{{- if $env -}}
{{- printf "beaesthetic-%s" $env -}}
{{- end -}}
{{- end -}}
{{- define "appointment.renderEnvConfig" -}}
{{- $environment := include "appointment.environment" . -}}
{{- $envConfig := toYaml .Values.envConfig -}}
{{- $envConfig = replace "${environment}" $environment $envConfig -}}
{{- tpl $envConfig . -}}
{{- end -}}

