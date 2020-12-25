package copier

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCopyModelToModel(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	var target originModel
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	assert.Nil(t, Instance(nil).From(origin).CopyTo(&target))
	assert.Equal(t, origin, target)
}

func TestCopyModelToProto(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
	}
	type TargetLocation struct {
		City      string
		Latitude  float32
		Longitude float32
	}
	type targetModel struct {
		Id       string
		Name     string
		Age      int
		Money    float64
		BirthDay string
		Location TargetLocation
	}

	var target targetModel
	originTime := time.Now()
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  originTime,
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	assert.Nil(t, Instance(nil).Install(RFC3339Convertor).From(origin).CopyTo(&target))
	assert.Equal(t, target, targetModel{
		Name:     "MockModel",
		Age:      21,
		Money:    7.7,
		BirthDay: originTime.Format(RFC3339Mili),
		Location: TargetLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	})
}

func TestCopySlice(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	var targets []*originModel
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}
	origins := []originModel{origin, origin}

	assert.Nil(t, Instance(nil).From(origins).CopyTo(&targets))
	assert.Equal(t, 2, len(targets))
	assert.Equal(t, origin, *targets[0])
	assert.Equal(t, origin, *targets[1])
}

func TestSkipExists(t *testing.T) {
	type OriginLocation struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name      string
		Age       int
		Money     float64
		BirthDay  time.Time
		IsDeleted bool
		Location  OriginLocation
		Detail    *OriginLocation
	}

	target := originModel{}
	origin := originModel{
		Name:      "MockModel",
		Age:       21,
		Money:     7.7,
		BirthDay:  time.Now(),
		IsDeleted: false,
		Location: OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
		Detail: &OriginLocation{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	target.Name = "I already have a name"
	assert.Nil(t, Instance(NewOption().SetOverwrite(false)).From(origin).CopyTo(&target))
	origin.Name = target.Name
	assert.Equal(t, origin, target)
}

func TestTransformerModelToProto(t *testing.T) {
	type Location struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	type originModel struct {
		Name     string
		BirthDay time.Time
		StoreId  string
	}

	type targetModel struct {
		Id         string
		TargetName string
		Name       string
		CreatedAt  string
		Location   *Location
	}

	var targets []targetModel
	locationMapper := map[string]*Location{
		"12306": &Location{
			City:      "ShangHai",
			Latitude:  234.123412,
			Longitude: 3423.43265,
		},
	}

	origins := []originModel{
		{
			Name:     "MockModel1",
			BirthDay: time.Now(),
			StoreId:  "12306",
		},
		{
			Name:     "MockModel2",
			BirthDay: time.Now(),
			StoreId:  "12345",
		},
	}
	assert.Nil(t, Instance(nil).
		RegisterTransformer(map[string]interface{}{
			"Location": func(storeId string) *Location {
				if location, ok := locationMapper[storeId]; ok {
					return location
				}
				return nil
			},
		}).
		RegisterResetDiffField([]DiffFieldPair{
			{Origin: "Name", Targets: []string{"TargetName", "Name"}},
			{Origin: "BirthDay", Targets: []string{"CreatedAt"}},
			{Origin: "StoreId", Targets: []string{"Location"}}},
		).Install(RFC3339Convertor).From(origins).CopyTo(&targets))

	assert.Equal(t, targetModel{
		TargetName: "MockModel1",
		Name:       "MockModel1",
		CreatedAt:  origins[0].BirthDay.Format(RFC3339Mili),
		Location:   locationMapper["12306"],
	}, targets[0])
	assert.Equal(t, targetModel{
		TargetName: "MockModel2",
		Name:       "MockModel2",
		CreatedAt:  origins[1].BirthDay.Format(RFC3339Mili),
		Location:   nil,
	}, targets[1])
}
