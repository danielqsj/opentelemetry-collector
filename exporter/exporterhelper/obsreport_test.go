// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporterhelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opencensus.io/metric"
	"go.opencensus.io/tag"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/obsreport/obsreporttest"
)

func TestExportEnqueueFailure(t *testing.T) {
	set, err := obsreporttest.SetupTelemetry()
	require.NoError(t, err)
	defer set.Shutdown(context.Background())

	exporter := config.NewComponentID("fakeExporter")

	insts := newInstruments(metric.NewRegistry())
	obsrep := newObsExporter(obsreport.ExporterSettings{
		Level:                  configtelemetry.LevelNormal,
		ExporterID:             exporter,
		ExporterCreateSettings: set.ToExporterCreateSettings(),
	}, insts)

	logRecords := int64(7)
	obsrep.recordLogsEnqueueFailure(context.Background(), logRecords)
	checkExporterEnqueueFailedLogsStats(t, insts, exporter, logRecords)

	spans := int64(12)
	obsrep.recordTracesEnqueueFailure(context.Background(), spans)
	checkExporterEnqueueFailedTracesStats(t, insts, exporter, spans)

	metricPoints := int64(21)
	obsrep.recordMetricsEnqueueFailure(context.Background(), metricPoints)
	checkExporterEnqueueFailedMetricsStats(t, insts, exporter, metricPoints)
}

// checkExporterEnqueueFailedTracesStats checks that reported number of spans failed to enqueue match given values.
// When this function is called it is required to also call SetupTelemetry as first thing.
func checkExporterEnqueueFailedTracesStats(t *testing.T, insts *instruments, exporter config.ComponentID, spans int64) {
	checkValueForProducer(t, insts.registry, tagsForExporterView(exporter), spans, "exporter/enqueue_failed_spans")
}

// checkExporterEnqueueFailedMetricsStats checks that reported number of metric points failed to enqueue match given values.
// When this function is called it is required to also call SetupTelemetry as first thing.
func checkExporterEnqueueFailedMetricsStats(t *testing.T, insts *instruments, exporter config.ComponentID, metricPoints int64) {
	checkValueForProducer(t, insts.registry, tagsForExporterView(exporter), metricPoints, "exporter/enqueue_failed_metric_points")
}

// checkExporterEnqueueFailedLogsStats checks that reported number of log records failed to enqueue match given values.
// When this function is called it is required to also call SetupTelemetry as first thing.
func checkExporterEnqueueFailedLogsStats(t *testing.T, insts *instruments, exporter config.ComponentID, logRecords int64) {
	checkValueForProducer(t, insts.registry, tagsForExporterView(exporter), logRecords, "exporter/enqueue_failed_log_records")
}

// tagsForExporterView returns the tags that are needed for the exporter views.
func tagsForExporterView(exporter config.ComponentID) []tag.Tag {
	return []tag.Tag{
		{Key: exporterTag, Value: exporter.String()},
	}
}
