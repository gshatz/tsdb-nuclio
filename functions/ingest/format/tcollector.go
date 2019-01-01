package format

import (
	"encoding/json"
	"github.com/nuclio/nuclio-sdk-go"
	"github.com/pkg/errors"
	"github.com/v3io/v3io-tsdb/pkg/tsdb"
	"strings"
)

/* sample event:
[{"timestamp": 1539967976, "metric": "net.sockstat.num_sockets", "value": 28.0, "tags": {"cati_id": "Unknown", "host": "nyl96i-9902", "type": "tcp", "envir": "Unknown"}},
 {"timestamp": 1539967976, "metric": "net.sockstat.num_timewait", "value": 0.0, "tags": {"cati_id": "Unknown", "host": "nyl96i-9902", "envir": "Unknown"}},
 {"timestamp": 1539967976, "metric": "net.sockstat.sockets_inuse", "value": 25.0, "tags": {"cati_id": "Unknown", "host": "nyl96i-9902", "type": "tcp", "envir": "Unknown"}},
 {"timestamp": 1539967976, "metric": "net.sockstat.sockets_inuse", "value": 12.0, "tags": {"cati_id": "Unknown", "host": "nyl96i-9902", "type": "udp", "envir": "Unknown"}}]
*/
type tInfo struct {
	Timestamp int64
	Metric    string
	Value     float64
	Tags      map[string]string
}

//implements InputFormat
type tcollectorFormat struct{}

func (Ingester tcollectorFormat) Ingest(tsdbAppender tsdb.Appender, event nuclio.Event) error {

	body := event.GetBody()
	tinfos := make([]tInfo, 0)

	// parse body
	if err := json.Unmarshal(body, &tinfos); err != nil {
		return errors.Wrapf(err, "Failed to parse request: %s", body)
	}

	for _, tinfo := range tinfos {

		metric := strings.Replace(tinfo.Metric, ".", "_", -1)

		sampleTime := tinfo.Timestamp * 1000
		sampleValue := tinfo.Value

		tagMap := make(map[string]string, 1000)
		for k, v := range tinfo.Tags {
			tagMap[k] = v
		}
		// convert the map[string]string -> []Labels
		labels := getLabelsFromRequest(metric, tagMap)

		_, err := tsdbAppender.Add(labels, sampleTime, sampleValue)
		if err != nil {
			return errors.Wrapf(err, "Failed to add samples. Request body is: %s", body)
		}

	}
	return nil
}
