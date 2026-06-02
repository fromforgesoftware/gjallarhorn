{{- define "gjallarhorn.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "gjallarhorn.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "gjallarhorn.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "gjallarhorn.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{ include "gjallarhorn.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: forge
{{- end -}}

{{- define "gjallarhorn.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gjallarhorn.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "gjallarhorn.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "gjallarhorn.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "gjallarhorn.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "gjallarhorn.dbSecretName" -}}
{{- if .Values.database.existingSecret -}}
{{- .Values.database.existingSecret -}}
{{- else -}}
{{- printf "%s-db" (include "gjallarhorn.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "gjallarhorn.channelSecretName" -}}
{{- if .Values.channels.existingSecret -}}
{{- .Values.channels.existingSecret -}}
{{- else -}}
{{- printf "%s-channels" (include "gjallarhorn.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/* Shared env block for server + migrator. */}}
{{- define "gjallarhorn.env" -}}
- name: SVC_NAME
  value: {{ include "gjallarhorn.name" . | quote }}
- name: REST_ADDRESS
  value: ":{{ .Values.ports.http }}"
- name: HTTP_ADDRESS
  value: ":{{ .Values.ports.http }}"
- name: GRPC_ADDRESS
  value: ":{{ .Values.ports.grpc }}"
- name: DB_HOST
  value: {{ .Values.database.host | quote }}
- name: DB_PORT
  value: {{ .Values.database.port | quote }}
- name: DB_NAME
  value: {{ .Values.database.name | quote }}
- name: DB_SCHEMA
  value: {{ .Values.database.schema | quote }}
- name: DB_SSL
  value: {{ .Values.database.ssl | quote }}
- name: DB_LOG_LEVEL
  value: {{ .Values.database.logLevel | default "warn" | quote }}
- name: DB_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.dbSecretName" . }}
      key: DB_USER
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.dbSecretName" . }}
      key: DB_PASSWORD
{{- if .Values.gatewaySecret }}
- name: FORGE_GATEWAY_SECRET
  value: {{ .Values.gatewaySecret | quote }}
{{- end }}
{{/* EMAIL channel (SMTP). Non-secret host/addr/from from values; password from secret. */}}
{{- with .Values.channels.email }}
- name: SMTP_ADDR
  value: {{ .addr | quote }}
- name: SMTP_HOST
  value: {{ .host | quote }}
- name: SMTP_FROM
  value: {{ .from | quote }}
- name: SMTP_USERNAME
  value: {{ .username | quote }}
{{- end }}
- name: SMTP_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.channelSecretName" . }}
      key: SMTP_PASSWORD
{{/* WEBHOOK channel: HMAC signing secret. */}}
- name: WEBHOOK_SIGNING_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.channelSecretName" . }}
      key: WEBHOOK_SIGNING_SECRET
{{/* SMS channel: HTTP gateway URL (non-secret) + auth header (secret). */}}
- name: SMS_ENDPOINT_URL
  value: {{ .Values.channels.sms.endpointUrl | quote }}
- name: SMS_AUTH_HEADER
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.channelSecretName" . }}
      key: SMS_AUTH_HEADER
{{/* PUSH channel: HTTP gateway URL (non-secret) + auth header (secret). */}}
- name: PUSH_ENDPOINT_URL
  value: {{ .Values.channels.push.endpointUrl | quote }}
- name: PUSH_AUTH_HEADER
  valueFrom:
    secretKeyRef:
      name: {{ include "gjallarhorn.channelSecretName" . }}
      key: PUSH_AUTH_HEADER
{{- end -}}
