{{- define "scheduler.environment" -}}
{{- default "" .Values.environment -}}
{{- end -}}

{{- define "scheduler.databaseName" -}}
{{- $base := .Values.postgres.database -}}
{{- $env := include "scheduler.environment" . -}}
{{- if $env -}}
{{- printf "%s_%s" $env $base -}}
{{- else -}}
{{- $base -}}
{{- end -}}
{{- end -}}

{{- define "scheduler.rabbitmqVhost" -}}
{{- $env := include "scheduler.environment" . -}}
{{- if $env -}}
{{- printf "beaesthetic-%s" $env -}}
{{- end -}}
{{- end -}}
