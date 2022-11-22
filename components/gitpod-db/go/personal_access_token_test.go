// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package db_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	db "github.com/gitpod-io/gitpod/components/gitpod-db/go"
	"github.com/gitpod-io/gitpod/components/gitpod-db/go/dbtest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPersonalAccessToken_Get(t *testing.T) {
	conn := dbtest.ConnectForTests(t)

	token := db.PersonalAccessToken{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Hash:           "some-secure-hash",
		Name:           "some-name",
		Description:    "some-description",
		Scopes:         []string{"read", "write"},
		ExpirationTime: time.Now().Add(5),
		CreatedAt:      time.Now(),
		LastModified:   time.Now(),
	}

	tx := conn.Create(token)
	require.NoError(t, tx.Error)

	result, err := db.GetToken(context.Background(), conn, token.ID)
	require.NoError(t, err)
	require.Equal(t, token.ID, result.ID)
}

func TestPersonalAccessToken_Create(t *testing.T) {
	conn := dbtest.ConnectForTests(t)

	request := db.PersonalAccessToken{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Hash:           "another-secure-hash",
		Name:           "another-name",
		Description:    "another-description",
		Scopes:         []string{"read", "write"},
		ExpirationTime: time.Now().Add(5),
		CreatedAt:      time.Now(),
		LastModified:   time.Now(),
	}

	result, err := db.CreatePersonalAccessToken(context.Background(), conn, request)
	require.NoError(t, err)

	require.Equal(t, request.ID, result.ID)
}

func TestListPersonalAccessTokensForUser(t *testing.T) {
	ctx := context.Background()
	conn := dbtest.ConnectForTests(t)
	pagination := db.Pagination{
		Page:     1,
		PageSize: 10,
	}

	userA := uuid.New()
	userB := uuid.New()

	now := time.Now().UTC()

	dbtest.CreatePersonalAccessTokenRecords(t, conn,
		dbtest.NewPersonalAccessToken(t, db.PersonalAccessToken{
			UserID:    userA,
			CreatedAt: now.Add(-1 * time.Minute),
		}),
		dbtest.NewPersonalAccessToken(t, db.PersonalAccessToken{
			UserID:    userA,
			CreatedAt: now,
		}),
		dbtest.NewPersonalAccessToken(t, db.PersonalAccessToken{
			UserID: userB,
		}),
	)

	tokensForUserA, err := db.ListPersonalAccessTokensForUser(ctx, conn, userA, pagination)
	require.NoError(t, err)
	require.Len(t, tokensForUserA.Results, 2)

	tokensForUserB, err := db.ListPersonalAccessTokensForUser(ctx, conn, userB, pagination)
	require.NoError(t, err)
	require.Len(t, tokensForUserB.Results, 1)

	tokensForUserWithNoData, err := db.ListPersonalAccessTokensForUser(ctx, conn, uuid.New(), pagination)
	require.NoError(t, err)
	require.Len(t, tokensForUserWithNoData.Results, 0)
}

func TestListPersonalAccessTokensForUser_PaginateThroughResults(t *testing.T) {
	ctx := context.Background()
	conn := dbtest.ConnectForTests(t).Debug()

	userA := uuid.New()

	total := 11
	var toCreate []db.PersonalAccessToken
	for i := 0; i < total; i++ {
		toCreate = append(toCreate, dbtest.NewPersonalAccessToken(t, db.PersonalAccessToken{
			UserID: userA,
			Name:   strconv.Itoa(i),
		}))
	}

	dbtest.CreatePersonalAccessTokenRecords(t, conn, toCreate...)

	batch1, err := db.ListPersonalAccessTokensForUser(ctx, conn, userA, db.Pagination{
		Page:     1,
		PageSize: 5,
	})
	require.NoError(t, err)
	require.Len(t, batch1.Results, 5)
	require.EqualValues(t, batch1.Total, total)

	batch2, err := db.ListPersonalAccessTokensForUser(ctx, conn, userA, db.Pagination{
		Page:     2,
		PageSize: 5,
	})
	require.NoError(t, err)
	require.Len(t, batch2.Results, 5)
	require.EqualValues(t, batch2.Total, total)

	batch3, err := db.ListPersonalAccessTokensForUser(ctx, conn, userA, db.Pagination{
		Page:     3,
		PageSize: 5,
	})
	require.NoError(t, err)
	require.Len(t, batch3.Results, 1)
	require.EqualValues(t, batch3.Total, total)
}