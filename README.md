# Copier

Copier for golang, copy value from struct to struct. Extends structmapper(https://github.com/structmapper/structmapper)

## Feature

- overwrite 支持在已有对象上增量复制，仅支持 struct to struct
- transformer 支持在复制过程中的自定义处理，支持多级
- reset different field name 支持不同名字段复制

## Usage

```go
type Location struct {
	City      string
	Latitude  float64
	Longitude float66
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
func TestTransformerModelToProto() {
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
	Instance(nil).
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
    fmt.Printf("targets %+v\n", targets)
    
    // Output:
    // targets &{ID:12345 Name:山田太郎 Age:32}
    // [{Id: TargetName:MockModel1 Name:MockModel1 CreatedAt:2020-12-29T13:55:17.883+08:00 Location:0xc00000e660} {Id: TargetName:MockModel2 Name:MockModel2 CreatedAt:2020-12-29T13:55:17.883+08:00 Location:<nil>}]
}
```

```go
type Age struct {
	Value int
}
type OriginCityInfo struct {
	Age  Age
	Area float64
}
type TargetCityInfo struct {
	Age  Age
	Name string
}
type OriginLocation struct {
	City     string
	CityInfo OriginCityInfo
}
type originModel struct {
	Name     string
	Location OriginLocation
}
type TargetLocation struct {
	City         string
	CityName     string
	CityNickName string
	CityInfo     TargetCityInfo
}
type targetModel struct {
	Name string
	Loc  *TargetLocation
}
func TestCopyModelToProtoWithMultiLevelAndTransformer() {
	var targets []targetModel
	origins := []originModel{
		{
			Name: "MockModel",
			Location: OriginLocation{
				City: "ShangHai",
				CityInfo: OriginCityInfo{
					Age:  Age{Value: 1},
					Area: 1,
				},
			},
		},
	}
    Instance(nil).RegisterTransformer(map[string]interface{}{
    		"Loc.CityNickName": func(city string) string {
    			return "Transformer city nick name"
    		},
    		"Loc.CityInfo.Age": func(city Age) Age {
    			city.Value++
    			return city
    		},
    	}).RegisterResetDiffField([]DiffFieldPair{
    		{Origin: "Location", Targets: []string{"Loc"}},
    		{Origin: "Location.City", Targets: []string{"Loc.CityName", "Loc.CityNickName"}},
    		{Origin: "Location.CityInfo.Age", Targets: []string{"Loc.CityInfo.Age"}},
    	}).From(origins).CopyTo(&targets)

}/done
```
