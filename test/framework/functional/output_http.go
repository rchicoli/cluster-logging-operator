package functional

import (
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional/common"
)

const (
	VectorHttpSourceConf = `
[sources.my_source]
type = "http"
address = "127.0.0.1:8090"
encoding = "ndjson"

[sinks.my_sink]
inputs = ["my_source"]
type = "file"
path = "/tmp/app-logs"

[sinks.my_sink.encoding]
codec = "ndjson"
`

	FluentdHttpSourceConf = `
<system>
  log_level debug
</system>
<source>
  @type http
  port 8090
  bind 0.0.0.0
  body_size_limit 32m
  keepalive_timeout 10s
</source>
# send fluentd logs to stdout
<match fluent.**>
  @type stdout
</match>
<match **>
  @type file
  append true
  path /tmp/app.logs
  symlink_path /tmp/app-logs
  <format>
    @type json
  </format>
</match>
`
)

func (f *CollectorFunctionalFramework) AddVectorHttpOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding vector http output", "name", output.Name)
	name := strings.ToLower(output.Name)

	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"vector.toml": VectorHttpSourceConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "vector.toml", VectorHttpSourceConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding vector container", "name", name)
	b.AddContainer(name, utils.GetComponentImage(constants.VectorName)).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		AddEnvVar("VECTOR_LOG", common.AdaptLogLevel()).
		AddEnvVar("VECTOR_INTERNAL_LOG_RATE_LIMIT", "0").
		WithCmd([]string{"vector", "--config-toml", "/tmp/config/vector.toml"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}

func (f *CollectorFunctionalFramework) AddFluentdHttpOutput(b *runtime.PodBuilder, output logging.OutputSpec) error {
	log.V(2).Info("Adding fluentd http output", "name", output.Name)
	name := strings.ToLower(output.Name)

	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"fluent.conf": FluentdHttpSourceConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "fluent.conf", VectorHttpSourceConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding fluentd container", "name", name)
	b.AddContainer(name, utils.GetComponentImage(constants.FluentdName)).
		AddVolumeMount(config.Name, "/tmp/config", "", false).
		WithCmd([]string{"fluentd", "-c", "/tmp/config/fluent.conf"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}
