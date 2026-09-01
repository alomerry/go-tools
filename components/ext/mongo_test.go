package ext

import (
	"bytes"
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/mongo"
	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MongoExtSuite 对 MongoExt 进行冒烟测试，连接真实的 MongoDB 实例。遵循项目的集成测试
// 约定（参见 components/mongo/mongo_test.go）：手动构造 ext 实例并注入真实的
// *mongo.Mongo 客户端，绕过 Init()，因此不依赖 apollo。当本机没有 MongoDB 时，
// 使用 `go test -short` 可跳过这些集成测试。
func TestMongoExtSuite(t *testing.T) {
	suite.Run(t, new(MongoExtSuite))
}

type MongoExtSuite struct {
	test.BaseSuite
	ext        *MongoExt
	dbName     string
	collection string
}

func (s *MongoExtSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping mongo integration test in short mode")
	}

	ctx := context.Background()
	s.dbName = "homelab"
	s.collection = "ext_test"

	cli, err := mongo.NewMongoClient(ctx, env.GetMongoDSN())
	if err != nil {
		s.T().Fatalf("connect mongo failed: %v", err)
	}

	// 手动构造 ext 实例并注入真实客户端。有意不调用 Init()：它会从 apollo
	// 拉取配置，而这对被测的 CRUD 方法而言无关紧要。
	s.ext = &MongoExt{
		cfg: &MongoExtConfig{Uri: env.GetMongoDSN()},
		cli: cli,
	}

	// 从一个干净的集合开始。
	if err := cli.Client().Database(s.dbName).Collection(s.collection).Drop(ctx); err != nil {
		s.T().Fatalf("drop collection failed: %v", err)
	}
}

func (s *MongoExtSuite) TearDownSuite() {
	if s.ext == nil || s.ext.cli == nil {
		return
	}
	ctx := context.Background()
	if err := s.ext.cli.Client().Database(s.dbName).Collection(s.collection).Drop(ctx); err != nil {
		s.T().Logf("drop collection failed: %v", err)
	}
	if err := s.ext.cli.Close(ctx); err != nil {
		s.T().Logf("close mongo client failed: %v", err)
	}
}

// TestInsert 验证 Insert 写入的文档可以被读回。
func (s *MongoExtSuite) TestInsert() {
	ctx := context.Background()
	id := bson.NewObjectID()
	doc := bson.M{"_id": id, "name": "insert-test"}

	err := s.ext.Insert(ctx, s.collection, doc)
	assert.NoError(s.T(), err)

	// 直接通过驱动读回，确认写入已落库。
	var got bson.M
	err = s.ext.cli.Client().Database(s.dbName).Collection(s.collection).FindOne(ctx, bson.M{"_id": id}).Decode(&got)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "insert-test", got["name"])
}

// TestFindOne 验证 FindOne 能解码此前插入的文档。
func (s *MongoExtSuite) TestFindOne() {
	ctx := context.Background()
	id := bson.NewObjectID()
	assert.NoError(s.T(), s.ext.Insert(ctx, s.collection, bson.M{"_id": id, "name": "find-one-test"}))

	var got bson.M
	err := s.ext.FindOne(ctx, s.collection, bson.M{"_id": id}, &got)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "find-one-test", got["name"])

	// 文档不存在时返回未找到错误，而非零值。
	var missing bson.M
	err = s.ext.FindOne(ctx, s.collection, bson.M{"_id": bson.NewObjectID()}, &missing)
	assert.Error(s.T(), err)
}

// TestCount 验证 Count 返回匹配文档的数量。
func (s *MongoExtSuite) TestCount() {
	ctx := context.Background()

	// 使用专属选择器，使计数不受其它同级测试插入文档的影响而保持确定。
	tag := "count-test"
	for i := 0; i < 3; i++ {
		assert.NoError(s.T(), s.ext.Insert(ctx, s.collection, bson.M{
			"_id":  bson.NewObjectID(),
			"name": "count-test",
			"tag":  tag,
		}))
	}

	total, err := s.ext.Count(ctx, s.collection, bson.M{"tag": tag})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)

	// 无匹配 -> 0，而非报错。
	zero, err := s.ext.Count(ctx, s.collection, bson.M{"tag": "nonexistent"})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), zero)
}

// TestFindPageWithSort 验证 FindPageWithSort 的显式排序生效，且空 sort 时
// 缺省行为与 FindPage（_id 升序）一致。
func (s *MongoExtSuite) TestFindPageWithSort() {
	ctx := context.Background()
	tag := "sort-test"

	for _, v := range []int{3, 1, 2} {
		assert.NoError(s.T(), s.ext.Insert(ctx, s.collection, bson.M{
			"_id": bson.NewObjectID(),
			"tag": tag,
			"seq": v,
		}))
	}

	// 显式 seq 降序。
	var desc []bson.M
	total, err := s.ext.FindPageWithSort(ctx, s.collection, bson.M{"tag": tag}, &desc, 1, 10, bson.D{{Key: "seq", Value: -1}})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), desc, 3)
	var gotDesc []int32
	for _, d := range desc {
		seq, ok := d["seq"].(int32)
		if assert.True(s.T(), ok, "seq should decode as int32, got %T", d["seq"]) {
			gotDesc = append(gotDesc, seq)
		}
	}
	assert.Equal(s.T(), []int32{3, 2, 1}, gotDesc)

	// 空 sort 缺省 _id 升序：除返回全集与总数外，还断言记录确实按 _id
	// 升序排列，守护缺省排序语义。
	var asc []bson.M
	total, err = s.ext.FindPageWithSort(ctx, s.collection, bson.M{"tag": tag}, &asc, 1, 10, nil)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), asc, 3)
	gotIDs := make([]bson.ObjectID, 0, len(asc))
	for _, d := range asc {
		id, ok := d["_id"].(bson.ObjectID)
		if assert.True(s.T(), ok, "_id should decode as ObjectID, got %T", d["_id"]) {
			gotIDs = append(gotIDs, id)
		}
	}
	for i := 1; i < len(gotIDs); i++ {
		assert.True(s.T(), bytes.Compare(gotIDs[i-1][:], gotIDs[i][:]) < 0,
			"records should be _id ascending: %v before %v", gotIDs[i-1], gotIDs[i])
	}

	// FindPage 保持原签名并委托给缺省排序。
	var viaFindPage []bson.M
	total, err = s.ext.FindPage(ctx, s.collection, bson.M{"tag": tag}, &viaFindPage, 1, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), viaFindPage, 3)
}

// TestFindPageClamp 固化 page/limit 非正数时的钳制行为：page<1 → 1、
// limit<1 → 10（缺省页大小），避免负数/零值产生非法 skip/limit 传给 mongo。
func (s *MongoExtSuite) TestFindPageClamp() {
	ctx := context.Background()
	tag := "clamp-test"

	for i := 0; i < 12; i++ {
		assert.NoError(s.T(), s.ext.Insert(ctx, s.collection, bson.M{
			"_id": bson.NewObjectID(),
			"tag": tag,
			"seq": i,
		}))
	}

	// 提取各文档的 seq（带 ok 的安全断言），用于比较页内容。
	seqsOf := func(docs []bson.M) []int32 {
		out := make([]int32, 0, len(docs))
		for _, d := range docs {
			v, ok := d["seq"].(int32)
			if assert.True(s.T(), ok, "seq should decode as int32, got %T", d["seq"]) {
				out = append(out, v)
			}
		}
		return out
	}

	// limit:-5 钳制为缺省页大小 10：12 条匹配只返回前 10 条。
	var clamped []bson.M
	total, err := s.ext.FindPageWithSort(ctx, s.collection, bson.M{"tag": tag}, &clamped, 0, -5, nil)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(12), total)
	assert.Len(s.T(), clamped, 10, "limit<1 should be clamped to default page size 10")
	assert.Equal(s.T(), []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, seqsOf(clamped))

	// page:0 钳制为 1：与显式 page=1 的同 limit 结果一致。
	var clampedPage []bson.M
	_, err = s.ext.FindPage(ctx, s.collection, bson.M{"tag": tag}, &clampedPage, 0, 5)
	assert.NoError(s.T(), err)
	var explicitPage []bson.M
	_, err = s.ext.FindPage(ctx, s.collection, bson.M{"tag": tag}, &explicitPage, 1, 5)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), seqsOf(explicitPage), seqsOf(clampedPage))
	assert.Equal(s.T(), []int32{0, 1, 2, 3, 4}, seqsOf(clampedPage))
}
