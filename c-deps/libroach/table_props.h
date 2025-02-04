// Copyright 2018 The Cockroach Authors.
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

#pragma once

#include <rocksdb/table_properties.h>

namespace cockroach {

// DBMakeTimeBoundCollector returns a TablePropertiesCollector hook to store the
// min and max MVCC timestamps present in each sstable in the metadata for that
// sstable. Used by the time bounded iterator optimization.
rocksdb::TablePropertiesCollectorFactory* DBMakeTimeBoundCollector();

// DBMakeDeleteRangeCollector returns a TablePropertiesCollector hook
// to mark sstables for compaction that contain range deletion
// tombstones. This ensures that range deletion tombstones are quickly
// compacted out of existence and the space for deleted data is
// reclaimed.
rocksdb::TablePropertiesCollectorFactory* DBMakeDeleteRangeCollector();

}  // namespace cockroach
