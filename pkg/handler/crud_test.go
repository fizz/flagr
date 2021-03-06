package handler

import (
	"testing"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/pkg/util"
	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

func TestCrudFlags(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	// step 0. it should get 0 flags when db is empty
	res = c.FindFlags(flag.FindFlagsParams{})
	assert.Len(t, res.(*flag.FindFlagsOK).Payload, 0)

	// step 1. it should be able to create one flag
	res = c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	assert.NotZero(t, res.(*flag.CreateFlagOK).Payload.ID)

	// step 2. it should be able to find some flags after creation
	res = c.FindFlags(flag.FindFlagsParams{})
	assert.NotZero(t, len(res.(*flag.FindFlagsOK).Payload))

	// step 3. it should be able to get the flag after creation
	res = c.GetFlag(flag.GetFlagParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.GetFlagOK).Payload.ID)

	// step 4. it should be able to put the flag
	res = c.PutFlag(flag.PutFlagParams{
		FlagID: int64(1),
		Body: &models.PutFlagRequest{
			Description: util.StringPtr("another funny flag"),
		}},
	)
	assert.NotZero(t, res.(*flag.PutFlagOK).Payload.ID)

	// step 5. it should be able to set the flag enabled state
	res = c.SetFlagEnabledState(flag.SetFlagEnabledParams{
		FlagID: int64(1),
		Body: &models.SetFlagEnabledRequest{
			Enabled: util.BoolPtr(true),
		}},
	)
	assert.True(t, *res.(*flag.SetFlagEnabledOK).Payload.Enabled)

	// step 6. it should be able to get the flag snapshot
	res = c.GetFlagSnapshots(flag.GetFlagSnapshotsParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.GetFlagSnapshotsOK).Payload)

	// step 7. it should be able to delete the flag
	res = c.DeleteFlag(flag.DeleteFlagParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.DeleteFlagOK))
}

func TestCrudFlagsWithFailures(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	// step 0. can't find non-exist flag
	res = c.GetFlag(flag.GetFlagParams{FlagID: int64(1)})
	assert.NotZero(t, res.(*flag.GetFlagDefault).Payload)
}

func TestCrudSegments(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})

	// step 1. it should be able to create segment
	res = c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	assert.NotZero(t, res.(*segment.CreateSegmentOK).Payload)
	res = c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment2"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	assert.NotZero(t, res.(*segment.CreateSegmentOK).Payload)

	// step 2. it should be able to find the segments
	res = c.FindSegments(segment.FindSegmentsParams{FlagID: int64(1)})
	assert.NotZero(t, len(res.(*segment.FindSegmentsOK).Payload))

	// step 3. it should be able to put the segment
	res = c.PutSegment(segment.PutSegmentParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(0)),
		},
	})
	assert.NotZero(t, res.(*segment.PutSegmentOK).Payload.ID)

	// step 4. it should be able to reorder the segments
	res = c.PutSegmentsReorder(segment.PutSegmentsReorderParams{
		FlagID: int64(1),
		Body: &models.PutSegmentReorderRequest{
			SegmentIds: []int64{int64(2), int64(1)},
		},
	})
	assert.NotZero(t, res.(*segment.PutSegmentsReorderOK))

	// step 5. it should have the correct order of segments
	res = c.FindSegments(segment.FindSegmentsParams{FlagID: int64(1)})
	assert.Equal(t, res.(*segment.FindSegmentsOK).Payload[0].ID, int64(2))

	// step 6. it should be able to delete the segment
	res = c.DeleteSegment(segment.DeleteSegmentParams{
		FlagID:    int64(1),
		SegmentID: int64(2),
	})
	assert.NotZero(t, res.(*segment.DeleteSegmentOK))
}

func TestCrudConstraints(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})

	// step 1. it should return 0 constraints before creaetion
	res = c.FindConstraints(constraint.FindConstraintsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.Zero(t, len(res.(*constraint.FindConstraintsOK).Payload))

	// step 2. it should be able to create a constraint
	res = c.CreateConstraint(constraint.CreateConstraintParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.CreateConstraintRequest{
			Operator: util.StringPtr("EQ"),
			Property: util.StringPtr("state"),
			Value:    util.StringPtr(`"NY"`),
		},
	})
	assert.NotZero(t, res.(*constraint.CreateConstraintOK).Payload.ID)

	// step 3. it should return some constraints when we get
	res = c.FindConstraints(constraint.FindConstraintsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.NotZero(t, len(res.(*constraint.FindConstraintsOK).Payload))

	// step 4. it should be able to put the constraint
	res = c.PutConstraint(constraint.PutConstraintParams{
		FlagID:       int64(1),
		SegmentID:    int64(1),
		ConstraintID: int64(1),
		Body: &models.CreateConstraintRequest{
			Operator: util.StringPtr("EQ"),
			Property: util.StringPtr("state"),
			Value:    util.StringPtr(`"CA"`),
		},
	})
	assert.NotZero(t, res.(*constraint.PutConstraintOK).Payload.ID)

	// step 5. it should be able to delete a constraint
	res = c.DeleteConstraint(constraint.DeleteConstraintParams{
		FlagID:       int64(1),
		SegmentID:    int64(1),
		ConstraintID: int64(1),
	})
	assert.NotZero(t, res.(*constraint.DeleteConstraintOK))
}

func TestCrudVariants(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})

	// step 0. it should return 0 variants before creaetion
	res = c.FindVariants(variant.FindVariantsParams{
		FlagID: int64(1),
	})
	assert.Zero(t, len(res.(*variant.FindVariantsOK).Payload))

	// step 1. it should be able to create variant
	res = c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})
	assert.NotZero(t, res.(*variant.CreateVariantOK).Payload.ID)

	// step 2. it should return some variants after creaetion
	res = c.FindVariants(variant.FindVariantsParams{
		FlagID: int64(1),
	})
	assert.NotZero(t, len(res.(*variant.FindVariantsOK).Payload))

	// step 3. it should be able to put variant
	res = c.PutVariant(variant.PutVariantParams{
		FlagID:    int64(1),
		VariantID: int64(1),
		Body: &models.PutVariantRequest{
			Key: util.StringPtr("another_control"),
		},
	})
	assert.Equal(t, *res.(*variant.PutVariantOK).Payload.Key, "another_control")

	// step 4. it should be able to delete the variant
	res = c.DeleteVariant(variant.DeleteVariantParams{
		FlagID:    int64(1),
		VariantID: int64(1),
	})
	assert.NotZero(t, res.(*variant.DeleteVariantOK))
}

func TestCrudDistributions(t *testing.T) {
	var res middleware.Responder
	db := entity.NewTestDB()
	c := &crud{}

	defer db.Close()
	defer gostub.StubFunc(&getDB, db).Reset()

	c.CreateFlag(flag.CreateFlagParams{
		Body: &models.CreateFlagRequest{
			Description: util.StringPtr("funny flag"),
		},
	})
	c.CreateSegment(segment.CreateSegmentParams{
		FlagID: int64(1),
		Body: &models.CreateSegmentRequest{
			Description:    util.StringPtr("segment1"),
			RolloutPercent: util.Int64Ptr(int64(100)),
		},
	})
	c.CreateVariant(variant.CreateVariantParams{
		FlagID: int64(1),
		Body: &models.CreateVariantRequest{
			Key: util.StringPtr("control"),
		},
	})

	// step 0. it should return 0 distributions before the creation
	res = c.FindDistributions(distribution.FindDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.Zero(t, len(res.(*distribution.FindDistributionsOK).Payload))

	// step 1. it should be able to create distribution
	res = c.PutDistributions(distribution.PutDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
		Body: &models.PutDistributionsRequest{
			Distributions: []*models.Distribution{
				{
					Percent:    util.Int64Ptr(int64(100)),
					VariantID:  util.Int64Ptr(int64(1)),
					VariantKey: util.StringPtr("control"),
				},
			},
		},
	})
	assert.NotZero(t, res.(*distribution.PutDistributionsOK).Payload)

	// step 2. it should return some distributions before the creation
	res = c.FindDistributions(distribution.FindDistributionsParams{
		FlagID:    int64(1),
		SegmentID: int64(1),
	})
	assert.NotZero(t, len(res.(*distribution.FindDistributionsOK).Payload))
}
