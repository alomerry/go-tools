package def

/*func TestWithEndpoint(t *testing.T) {
	t.Run("sets endpoint", func(t *testing.T) {
		meta := &Meta{}
		WithEndpoint("http://localhost:8086")(meta)
		assert.Equal(t, "http://localhost:8086", meta.Endpoint)
	})
}

func TestWithOrg(t *testing.T) {
	t.Run("sets org", func(t *testing.T) {
		meta := &Meta{}
		WithOrg("test-org")(meta)
		assert.Equal(t, "test-org", meta.Org)
	})
}

func TestWithToken(t *testing.T) {
	t.Run("sets token", func(t *testing.T) {
		meta := &Meta{}
		WithToken("test-token")(meta)
		assert.Equal(t, "test-token", meta.Token)
	})
}

func TestWithBucket(t *testing.T) {
	t.Run("sets bucket", func(t *testing.T) {
		meta := &Meta{}
		WithBucket("test-bucket")(meta)
		assert.Equal(t, "test-bucket", meta.Bucket)
	})
}

func TestNewPointBuilder(t *testing.T) {
	t.Run("creates builder with measurement", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		assert.NotNil(t, builder)
		assert.Equal(t, "temperature", builder.measurement)
		assert.NotNil(t, builder.tags)
		assert.NotNil(t, builder.fields)
	})
}

func TestPointBuilder_Tag(t *testing.T) {
	t.Run("adds single tag", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Tag("sensor", "A1")

		point := builder.Build()
		assert.Equal(t, "A1", point.Tags["sensor"])
	})

	t.Run("returns builder for chaining", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		result := builder.Tag("sensor", "A1")
		assert.Equal(t, builder, result)
	})

	t.Run("overwrites existing tag", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Tag("sensor", "A1")
		builder.Tag("sensor", "A2")

		point := builder.Build()
		assert.Equal(t, "A2", point.Tags["sensor"])
	})
}

func TestPointBuilder_Tags(t *testing.T) {
	t.Run("adds multiple tags", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Tags(map[string]string{
			"sensor":   "A1",
			"location": "room1",
		})

		point := builder.Build()
		assert.Equal(t, "A1", point.Tags["sensor"])
		assert.Equal(t, "room1", point.Tags["location"])
	})

	t.Run("merges with existing tags", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Tag("sensor", "A1")
		builder.Tags(map[string]string{
			"location": "room1",
		})

		point := builder.Build()
		assert.Equal(t, "A1", point.Tags["sensor"])
		assert.Equal(t, "room1", point.Tags["location"])
	})
}

func TestPointBuilder_Field(t *testing.T) {
	t.Run("adds single field", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Field("value", 23.5)

		point := builder.Build()
		assert.Equal(t, 23.5, point.Fields["value"])
	})

	t.Run("adds different field types", func(t *testing.T) {
		builder := NewPointBuilder("test")
		builder.Field("int", 42)
		builder.Field("float", 3.14)
		builder.Field("string", "test")
		builder.Field("bool", true)

		point := builder.Build()
		assert.Equal(t, 42, point.Fields["int"])
		assert.Equal(t, 3.14, point.Fields["float"])
		assert.Equal(t, "test", point.Fields["string"])
		assert.Equal(t, true, point.Fields["bool"])
	})

	t.Run("overwrites existing field", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Field("value", 23.5)
		builder.Field("value", 24.0)

		point := builder.Build()
		assert.Equal(t, 24.0, point.Fields["value"])
	})
}

func TestPointBuilder_Fields(t *testing.T) {
	t.Run("adds multiple fields", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Fields(map[string]any{
			"value": 23.5,
			"unit":  "celsius",
		})

		point := builder.Build()
		assert.Equal(t, 23.5, point.Fields["value"])
		assert.Equal(t, "celsius", point.Fields["unit"])
	})

	t.Run("merges with existing fields", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		builder.Field("value", 23.5)
		builder.Fields(map[string]any{
			"unit": "celsius",
		})

		point := builder.Build()
		assert.Equal(t, 23.5, point.Fields["value"])
		assert.Equal(t, "celsius", point.Fields["unit"])
	})
}

func TestPointBuilder_Time(t *testing.T) {
	t.Run("sets custom time", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		builder := NewPointBuilder("temperature")
		builder.Time(customTime)

		point := builder.Build()
		assert.Equal(t, customTime, point.Time)
	})

	t.Run("defaults to current time", func(t *testing.T) {
		before := time.Now()
		builder := NewPointBuilder("temperature")
		point := builder.Build()
		after := time.Now()

		assert.True(t, point.Time.After(before) || point.Time.Equal(before))
		assert.True(t, point.Time.Before(after) || point.Time.Equal(after))
	})
}

func TestPointBuilder_Build(t *testing.T) {
	t.Run("builds complete point", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		point := NewPointBuilder("temperature").
			Tag("sensor", "A1").
			Tag("location", "room1").
			Field("value", 23.5).
			Field("unit", "celsius").
			Time(customTime).
			Build()

		assert.Equal(t, "temperature", point.Measurement)
		assert.Equal(t, "A1", point.Tags["sensor"])
		assert.Equal(t, "room1", point.Tags["location"])
		assert.Equal(t, 23.5, point.Fields["value"])
		assert.Equal(t, "celsius", point.Fields["unit"])
		assert.Equal(t, customTime, point.Time)
	})

	t.Run("builds point with chaining", func(t *testing.T) {
		point := NewPointBuilder("cpu_usage").
			Tags(map[string]string{"host": "server1", "env": "prod"}).
			Fields(map[string]any{"usage": 45.2, "cores": 8}).
			Build()

		assert.Equal(t, "cpu_usage", point.Measurement)
		assert.Len(t, point.Tags, 2)
		assert.Len(t, point.Fields, 2)
	})
}

func TestPointBuilder_Chaining(t *testing.T) {
	t.Run("supports method chaining", func(t *testing.T) {
		point := NewPointBuilder("test").
			Tag("tag1", "value1").
			Tag("tag2", "value2").
			Field("field1", 1).
			Field("field2", 2.0).
			Tags(map[string]string{"tag3": "value3"}).
			Fields(map[string]any{"field3": "value3"}).
			Time(time.Now()).
			Build()

		assert.Equal(t, "test", point.Measurement)
		assert.Len(t, point.Tags, 3)
		assert.Len(t, point.Fields, 3)
	})
}
*/
