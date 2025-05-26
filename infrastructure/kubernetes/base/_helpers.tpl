{{/* Standard resource labels */}}
{{- define "phoenix.labels" -}}
app.kubernetes.io/name: {{ .name | default "phoenix" }}
app.kubernetes.io/instance: {{ .instance | default "default" }}
app.kubernetes.io/part-of: phoenix-platform
app.kubernetes.io/managed-by: {{ .manager | default "kustomize" }}
{{- if .component }}
app.kubernetes.io/component: {{ .component }}
{{- end }}
{{- end -}}

{{/* Standard resource limits */}}
{{- define "phoenix.resourceLimits" -}}
resources:
  limits:
    cpu: {{ .limits.cpu | default "500m" }}
    memory: {{ .limits.memory | default "512Mi" }}
  requests:
    cpu: {{ .requests.cpu | default "100m" }}
    memory: {{ .requests.memory | default "128Mi" }}
{{- end -}}

{{/* Standard security context */}}
{{- define "phoenix.securityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
{{- end -}}

{{/* Standard probes */}}
{{- define "phoenix.probes" -}}
{{- if and .http.path .http.port }}
livenessProbe:
  httpGet:
    path: {{ .http.path }}
    port: {{ .http.port }}
  initialDelaySeconds: {{ .liveness.initialDelay | default 30 }}
  periodSeconds: {{ .liveness.period | default 10 }}
readinessProbe:
  httpGet:
    path: {{ .http.path }}
    port: {{ .http.port }}
  initialDelaySeconds: {{ .readiness.initialDelay | default 5 }}
  periodSeconds: {{ .readiness.period | default 10 }}
{{- else if and .tcp.port }}
livenessProbe:
  tcpSocket:
    port: {{ .tcp.port }}
  initialDelaySeconds: {{ .liveness.initialDelay | default 30 }}
  periodSeconds: {{ .liveness.period | default 10 }}
readinessProbe:
  tcpSocket:
    port: {{ .tcp.port }}
  initialDelaySeconds: {{ .readiness.initialDelay | default 5 }}
  periodSeconds: {{ .readiness.period | default 10 }}
{{- end -}}
{{- end -}}
