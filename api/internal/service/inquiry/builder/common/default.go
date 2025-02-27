package common

import (
	"fmt"

	"github.com/clickvisual/clickvisual/api/internal/service/inquiry/builder/bumo"
)

func BuilderFieldsData(mapping string) string {
	if mapping == "" {
		mapping = `_source_ String,
  _cluster_ String,
  _log_agent_ String,
  _namespace_ String,
  _node_name_ String,
  _node_ip_ String,
  _container_name_ String,
  _pod_name_ String,`
	}
	return fmt.Sprintf(`(
  %s
  _time_second_ DateTime,
  _time_nanosecond_ DateTime64(9, 'Asia/Shanghai'),
  _raw_log_ String
)
`, mapping)
}

func BuilderFieldsStream(mapping, timeField, timeTyp, logField string) string {
	if timeField == "" {
		timeField = "_time_"
	}
	if logField == "" {
		logField = "_log_"
	}
	if mapping == "" {
		mapping = `_source_ String,
  _cluster_ String,
  _log_agent_ String,
  _namespace_ String,
  _node_name_ String,
  _node_ip_ String,
  _container_name_ String,
  _pod_name_ String,`
	}
	return fmt.Sprintf(`(
  %s
  %s %s,
  %s String
)
`, mapping, timeField, timeTyp, logField)
}

func BuilderFieldsView(mapping, logField string, paramsView bumo.ParamsView) string {
	if logField == "" {
		logField = "_log_"
	}
	if mapping == "" {
		mapping = `_source_,
  _cluster_,
  _log_agent_,
  _namespace_,
  _node_name_,
  _node_ip_,
  _container_name_,
  _pod_name_,`
	}
	return fmt.Sprintf(`SELECT
  %s
  %s,
  %s AS _raw_log_%s
FROM %s
`,
		mapping, paramsView.TimeConvert, logField, paramsView.CommonFields, paramsView.SourceTable)
}
