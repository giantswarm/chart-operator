package e2etemplates

const ApiextensionsChartConfigE2EChartValues = `chart:
  channel: {{ .Channel }}
  configMap:
    name: {{ .ConfigMapName }}
    namespace: {{ .ConfigMapNamespace }}
    resourceVersion: {{ .ConfigMapResourceVersion }}
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  release: {{ .Release }}
  secret:
    name: {{ .SecretName }}
    namespace: {{ .SecretNamespace }}
    resourceVersion: {{ .SecretResourceVersion }}
versionBundleVersion: {{ .VersionBundleVersion }}
`
