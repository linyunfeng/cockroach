// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package log

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/caller"
	"github.com/cockroachdb/cockroach/pkg/util/log/logtags"
)

// formatTags appends the tags to a strings.Builder. If there are no tags,
// returns false.
func formatTags(ctx context.Context, buf *strings.Builder) bool {
	tags := logtags.FromContext(ctx)
	if tags == nil {
		return false
	}
	buf.WriteByte('[')
	for i, t := range tags.Get() {
		if i > 0 {
			buf.WriteByte(',')
		}

		buf.WriteString(t.Key())
		if v := t.Value(); v != nil && v != "" {
			// For tags that have a value and are longer than a character,
			// we output "tag=value". For one character tags we don't use a
			// separator (e.g. "n1").
			if len(t.Key()) > 1 {
				buf.WriteByte('=')
			}
			fmt.Fprint(buf, v)
		}
	}
	buf.WriteString("] ")
	return true
}

// MakeMessage creates a structured log entry.
func MakeMessage(ctx context.Context, format string, args []interface{}) string {
	var buf strings.Builder
	formatTags(ctx, &buf)
	if len(args) == 0 {
		buf.WriteString(format)
	} else if len(format) == 0 {
		fmt.Fprint(&buf, args...)
	} else {
		fmt.Fprintf(&buf, format, args...)
	}
	return buf.String()
}

// addStructured creates a structured log entry to be written to the
// specified facility of the logger.
func addStructured(ctx context.Context, s Severity, depth int, format string, args []interface{}) {
	file, line, _ := caller.Lookup(depth + 1)
	msg := MakeMessage(ctx, format, args)

	if s == Severity_FATAL {
		// We load the ReportingSettings from the a global singleton in this
		// call path. See the singleton's comment for a rationale.
		if sv := settings.TODO(); sv != nil {
			SendCrashReport(ctx, sv, depth+2, format, args, ReportTypePanic)
		}
	}
	// MakeMessage already added the tags when forming msg, we don't want
	// eventInternal to prepend them again.
	eventInternal(ctx, (s >= Severity_ERROR), false /*withTags*/, "%s:%d %s", file, line, msg)
	logging.outputLogEntry(s, file, line, msg)
}
