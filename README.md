# Copier

Copier for golang, copy value from struct to struct. Extends structmapper(https://github.com/structmapper/structmapper)

## Feature

- overwrite 支持在已有对象上增量复制，仅支持 struct to struct
- transformer 支持在复制过程中的自定义处理
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
