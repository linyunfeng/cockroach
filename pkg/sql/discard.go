// Copyright 2017 The Cockroach Authors.
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

package sql

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/errors"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/pgcode"
)

// Discard implements the DISCARD statement.
// See https://www.postgresql.org/docs/9.6/static/sql-discard.html for details.
func (p *planner) Discard(ctx context.Context, s *tree.Discard) (planNode, error) {
	switch s.Mode {
	case tree.DiscardModeAll:
		if !p.autoCommit {
			return nil, pgerror.New(pgcode.ActiveSQLTransaction,
				"DISCARD ALL cannot run inside a transaction block")
		}

		// RESET ALL
		if err := resetSessionVars(ctx, p.sessionDataMutator); err != nil {
			return nil, err
		}

		// DEALLOCATE ALL
		p.preparedStatements.DeleteAll(ctx)
	default:
		return nil, errors.AssertionFailedf("unknown mode for DISCARD: %d", s.Mode)
	}
	return newZeroNode(nil /* columns */), nil
}

func resetSessionVars(ctx context.Context, m *sessionDataMutator) error {
	for _, varName := range varNames {
		v := varGen[varName]
		if v.Set != nil {
			hasDefault, defVal := getSessionVarDefaultString(varName, v, m)
			if hasDefault {
				if err := v.Set(ctx, m, defVal); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
