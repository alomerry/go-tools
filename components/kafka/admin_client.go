package kafka

// AdminClient 面向运维场景的 Kafka 管理客户端：topic 列表/元数据、offset 水位、
// consumer group 列表与 lag、消息采样，以及 topic/consumer group 的增删。
// 与 Client/Producer 同构，复用 Option 模式（WithAddress/WithSCRAMSASL/...），
// 基于 kafka-go 的协议层 API（Metadata/ListOffsets/OffsetFetch/DescribeGroups 等）。
import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type AdminClient struct {
	client    *kafka.Client
	transport *kafka.Transport
	dialer    *kafka.Dialer
	addresses []string
}

// NewAdminClient 创建管理客户端。与 NewDefaultProducer 一致：以 SCRAM SASL 走
// Transport/Dialer，探测类接口均接受 ctx（由调用方控制超时）。
func NewAdminClient(opts ...Option) (*AdminClient, error) {
	options := newOptions(opts...)
	if len(options.addresses) == 0 {
		return nil, errors.New("kafka admin client: missing broker address")
	}

	dialer := &net.Dialer{
		Timeout: options.writeTimeout,
	}
	transport := &kafka.Transport{
		DialTimeout: options.writeTimeout,
		Dial:        dialer.DialContext,
	}
	if options.sasl != nil {
		transport.SASL = options.sasl
	}

	readerDialer := &kafka.Dialer{
		Timeout:   30 * time.Second, // 与 NewKafkaDialer 保持一致（兼容 SSH 隧道/nginx 代理）
		KeepAlive: 10 * time.Second,
	}
	if options.sasl != nil {
		readerDialer.SASLMechanism = options.sasl
	}

	return &AdminClient{
		client: &kafka.Client{
			Addr:      kafka.TCP(options.addresses...),
			Timeout:   options.readTimeout,
			Transport: transport,
		},
		transport: transport,
		dialer:    readerDialer,
		addresses: options.addresses,
	}, nil
}

// Close 关闭底层连接池的空闲连接，可重复调用。
func (a *AdminClient) Close() error {
	if a.transport != nil {
		a.transport.CloseIdleConnections()
	}
	return nil
}

// TopicMeta 描述单个 topic 的元数据与运维指标。
type TopicMeta struct {
	Name              string
	Partitions        int
	ReplicationFactor int
	Internal          bool
	RetentionMs       int64
	SegmentBytes      int64
	// MessageCount 为消息量估算：各分区 logEndOffset-logStartOffset 求和。
	MessageCount int64
	// UnderReplicated 是否存在 ISR 数小于副本数的分区（存在未同步副本）。
	UnderReplicated bool
}

// ListTopics 返回全部 topic 的元数据。内部依次调用 Metadata（分区/副本/ISR）、
// DescribeConfigs（retention.ms/segment.bytes）、ListOffsets（消息量估算），
// 均为批量 API，一次往返完成。
func (a *AdminClient) ListTopics(ctx context.Context) ([]TopicMeta, error) {
	md, err := a.client.Metadata(ctx, &kafka.MetadataRequest{})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list topics failed: %w", err)
	}

	names := make([]string, 0, len(md.Topics))
	byName := make(map[string]kafka.Topic, len(md.Topics))
	for _, t := range md.Topics {
		byName[t.Name] = t
		names = append(names, t.Name)
	}
	if len(names) == 0 {
		return []TopicMeta{}, nil
	}

	configs, err := a.topicConfigs(ctx, names)
	if err != nil {
		return nil, err
	}
	counts, err := a.messageCounts(ctx, md.Topics)
	if err != nil {
		return nil, err
	}

	topics := make([]TopicMeta, 0, len(md.Topics))
	for _, name := range names {
		t := byName[name]
		meta := TopicMeta{
			Name:              t.Name,
			Partitions:        len(t.Partitions),
			ReplicationFactor: replicationFactor(t),
			Internal:          t.Internal,
			MessageCount:      counts[name],
			UnderReplicated:   underReplicated(t),
		}
		if entries := configs[name]; entries != nil {
			meta.RetentionMs, _ = strconv.ParseInt(entries["retention.ms"], 10, 64)
			meta.SegmentBytes, _ = strconv.ParseInt(entries["segment.bytes"], 10, 64)
		}
		topics = append(topics, meta)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].Name < topics[j].Name })
	return topics, nil
}

// topicConfigs 批量拉取各 topic 的配置项（无 ConfigNames 表示返回全部配置）。
func (a *AdminClient) topicConfigs(ctx context.Context, topics []string) (map[string]map[string]string, error) {
	resources := make([]kafka.DescribeConfigRequestResource, 0, len(topics))
	for _, name := range topics {
		resources = append(resources, kafka.DescribeConfigRequestResource{
			ResourceType: kafka.ResourceTypeTopic,
			ResourceName: name,
		})
	}
	resp, err := a.client.DescribeConfigs(ctx, &kafka.DescribeConfigsRequest{Resources: resources})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: describe configs failed: %w", err)
	}

	out := make(map[string]map[string]string, len(resp.Resources))
	for _, r := range resp.Resources {
		entries := make(map[string]string, len(r.ConfigEntries))
		for _, e := range r.ConfigEntries {
			entries[e.ConfigName] = e.ConfigValue
		}
		out[r.ResourceName] = entries
	}
	return out, nil
}

// messageCounts 估算各 topic 的消息量：分区 logEndOffset 与 logStartOffset 差值求和。
func (a *AdminClient) messageCounts(ctx context.Context, topics []kafka.Topic) (map[string]int64, error) {
	reqTopics := make(map[string][]kafka.OffsetRequest)
	for _, t := range topics {
		for _, p := range t.Partitions {
			reqTopics[t.Name] = append(reqTopics[t.Name], kafka.FirstOffsetOf(p.ID), kafka.LastOffsetOf(p.ID))
		}
	}
	resp, err := a.client.ListOffsets(ctx, &kafka.ListOffsetsRequest{Topics: reqTopics})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list offsets failed: %w", err)
	}

	counts := make(map[string]int64, len(topics))
	for topic, parts := range resp.Topics {
		var total int64
		for _, p := range parts {
			if p.FirstOffset >= 0 && p.LastOffset >= 0 {
				total += p.LastOffset - p.FirstOffset
			}
		}
		counts[topic] = total
	}
	return counts, nil
}

// PartitionOffset 描述单个分区的 offset 水位。
type PartitionOffset struct {
	Partition      int
	Leader         int
	HighWatermark  int64
	LogStartOffset int64
}

// TopicOffsets 返回各 topic 每个分区的 leader/highWatermark/logStartOffset。
// topics 为空表示全部 topic。
func (a *AdminClient) TopicOffsets(ctx context.Context, topics []string) (map[string][]PartitionOffset, error) {
	md, err := a.client.Metadata(ctx, &kafka.MetadataRequest{Topics: nilIfEmpty(topics)})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: topic metadata failed: %w", err)
	}

	reqTopics := make(map[string][]kafka.OffsetRequest)
	for _, t := range md.Topics {
		for _, p := range t.Partitions {
			reqTopics[t.Name] = append(reqTopics[t.Name], kafka.FirstOffsetOf(p.ID), kafka.LastOffsetOf(p.ID))
		}
	}
	offResp, err := a.client.ListOffsets(ctx, &kafka.ListOffsetsRequest{Topics: reqTopics})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list offsets failed: %w", err)
	}

	result := make(map[string][]PartitionOffset, len(md.Topics))
	for _, t := range md.Topics {
		offMap := make(map[int]kafka.PartitionOffsets, len(offResp.Topics[t.Name]))
		for _, po := range offResp.Topics[t.Name] {
			offMap[po.Partition] = po
		}
		parts := make([]PartitionOffset, 0, len(t.Partitions))
		for _, p := range t.Partitions {
			po, ok := offMap[p.ID]
			if !ok {
				continue
			}
			parts = append(parts, PartitionOffset{
				Partition:      p.ID,
				Leader:         p.Leader.ID,
				HighWatermark:  po.LastOffset,
				LogStartOffset: po.FirstOffset,
			})
		}
		result[t.Name] = parts
	}
	return result, nil
}

// ConsumerGroupMeta 描述单个消费组的基础信息。
type ConsumerGroupMeta struct {
	GroupID string
	State   string
	Members int
	// Protocol 为组协议类型（如 "consumer"），来自 ListGroups 响应。
	Protocol string
}

// ListConsumerGroups 返回全部消费组的基础信息（组名/状态/成员数/协议类型）。
func (a *AdminClient) ListConsumerGroups(ctx context.Context) ([]ConsumerGroupMeta, error) {
	listResp, err := a.client.ListGroups(ctx, &kafka.ListGroupsRequest{})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list groups failed: %w", err)
	}
	if len(listResp.Groups) == 0 {
		return []ConsumerGroupMeta{}, nil
	}

	ids := make([]string, 0, len(listResp.Groups))
	protocols := make(map[string]string, len(listResp.Groups))
	for _, g := range listResp.Groups {
		ids = append(ids, g.GroupID)
		protocols[g.GroupID] = g.ProtocolType
	}

	descResp, err := a.client.DescribeGroups(ctx, &kafka.DescribeGroupsRequest{GroupIDs: ids})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: describe groups failed: %w", err)
	}

	groups := make([]ConsumerGroupMeta, 0, len(descResp.Groups))
	for _, g := range descResp.Groups {
		groups = append(groups, ConsumerGroupMeta{
			GroupID:  g.GroupID,
			State:    g.GroupState,
			Members:  len(g.Members),
			Protocol: protocols[g.GroupID],
		})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].GroupID < groups[j].GroupID })
	return groups, nil
}

// PartitionLag 描述单个分区在当前消费组下的滞后情况。
type PartitionLag struct {
	Topic         string
	Partition     int
	CurrentOffset int64
	LogEndOffset  int64
	Lag           int64
}

// ConsumerGroupLag 描述一个消费组的滞后聚合：totalLag 为各分区 lag 之和。
type ConsumerGroupLag struct {
	GroupID    string
	TotalLag   int64
	Partitions []PartitionLag
}

// ConsumerGroupLag 计算指定消费组在全部 topic 分区上的 lag
// （logEndOffset-currentOffset 求和）。分区尚未提交 offset 时
// CurrentOffset 记 0，lag 为该分区全量消息数。
func (a *AdminClient) ConsumerGroupLag(ctx context.Context, group string) (*ConsumerGroupLag, error) {
	offResp, err := a.client.OffsetFetch(ctx, &kafka.OffsetFetchRequest{GroupID: group})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: fetch group offsets failed: %w", err)
	}
	if offResp.Error != nil {
		return nil, offResp.Error
	}
	if len(offResp.Topics) == 0 {
		return &ConsumerGroupLag{GroupID: group, Partitions: []PartitionLag{}}, nil
	}

	reqTopics := make(map[string][]kafka.OffsetRequest, len(offResp.Topics))
	for topic, parts := range offResp.Topics {
		for _, p := range parts {
			reqTopics[topic] = append(reqTopics[topic], kafka.LastOffsetOf(p.Partition))
		}
	}
	endResp, err := a.client.ListOffsets(ctx, &kafka.ListOffsetsRequest{Topics: reqTopics})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list offsets failed: %w", err)
	}
	endOffsets := make(map[string]map[int]int64)
	for topic, parts := range endResp.Topics {
		m := make(map[int]int64, len(parts))
		for _, p := range parts {
			m[p.Partition] = p.LastOffset
		}
		endOffsets[topic] = m
	}

	lag := &ConsumerGroupLag{GroupID: group, Partitions: []PartitionLag{}}
	for topic, parts := range offResp.Topics {
		for _, p := range parts {
			current := p.CommittedOffset
			if current < 0 {
				current = 0
			}
			logEnd := endOffsets[topic][p.Partition]
			if logEnd < 0 {
				logEnd = 0
			}
			l := logEnd - current
			if l < 0 {
				l = 0
			}
			lag.TotalLag += l
			lag.Partitions = append(lag.Partitions, PartitionLag{
				Topic:         topic,
				Partition:     p.Partition,
				CurrentOffset: current,
				LogEndOffset:  logEnd,
				Lag:           l,
			})
		}
	}
	sort.Slice(lag.Partitions, func(i, j int) bool {
		if lag.Partitions[i].Topic != lag.Partitions[j].Topic {
			return lag.Partitions[i].Topic < lag.Partitions[j].Topic
		}
		return lag.Partitions[i].Partition < lag.Partitions[j].Partition
	})
	return lag, nil
}

// SampleMessage 是消息采样返回的单条消息。
type SampleMessage struct {
	Partition int
	Offset    int64
	Key       string
	Timestamp time.Time
	Value     []byte
}

// ReadLatestMessages 从各分区最新 offset 往前回读最多 count 条消息，汇总后按
// 时间倒序（同时间按 offset 倒序）取最新的 count 条。适用于运维侧抽样查看最新
// 消息，不做跨分区的严格全局排序保证。
func (a *AdminClient) ReadLatestMessages(ctx context.Context, topic string, count int) ([]SampleMessage, error) {
	if count <= 0 {
		return nil, errors.New("kafka admin client: count must be positive")
	}

	md, err := a.client.Metadata(ctx, &kafka.MetadataRequest{Topics: []string{topic}})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: topic metadata failed: %w", err)
	}
	if len(md.Topics) == 0 {
		return nil, fmt.Errorf("kafka admin client: topic %q not found", topic)
	}
	parts := md.Topics[0].Partitions

	reqTopics := make(map[string][]kafka.OffsetRequest, len(parts))
	for _, p := range parts {
		reqTopics[topic] = append(reqTopics[topic], kafka.FirstOffsetOf(p.ID), kafka.LastOffsetOf(p.ID))
	}
	offResp, err := a.client.ListOffsets(ctx, &kafka.ListOffsetsRequest{Topics: reqTopics})
	if err != nil {
		return nil, fmt.Errorf("kafka admin client: list offsets failed: %w", err)
	}
	bounds := make(map[int]kafka.PartitionOffsets, len(parts))
	for _, po := range offResp.Topics[topic] {
		bounds[po.Partition] = po
	}

	var samples []SampleMessage
	for _, p := range parts {
		po, ok := bounds[p.ID]
		if !ok || po.LastOffset <= 0 {
			continue // 分区无数据（或未返回水位）
		}
		start := po.LastOffset - int64(count)
		if start < 0 {
			start = 0
		}
		// FirstOffset 未知（<0）时仅按 0 兜底；已知且大于 start 时（retention 已
		// 清理过老数据）从 logStartOffset 起读。
		if po.FirstOffset > 0 && start < po.FirstOffset {
			start = po.FirstOffset
		}

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:          a.addresses,
			Topic:            topic,
			Partition:        p.ID,
			Dialer:           a.dialer,
			MinBytes:         1,
			MaxBytes:         4 << 20,
			MaxWait:          500 * time.Millisecond,
			ReadBatchTimeout: 500 * time.Millisecond,
		})
		_ = reader.SetOffset(start)
		for i := 0; i < count; i++ {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				break // ctx 超时 / 无更多消息
			}
			samples = append(samples, SampleMessage{
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       string(msg.Key),
				Timestamp: msg.Time,
				Value:     msg.Value,
			})
		}
		_ = reader.Close()
	}

	sort.Slice(samples, func(i, j int) bool {
		if !samples[i].Timestamp.Equal(samples[j].Timestamp) {
			return samples[i].Timestamp.After(samples[j].Timestamp)
		}
		if samples[i].Offset != samples[j].Offset {
			return samples[i].Offset > samples[j].Offset
		}
		return samples[i].Partition < samples[j].Partition
	})
	if len(samples) > count {
		samples = samples[:count]
	}
	return samples, nil
}

// CreateTopic 创建 topic。configEntries 可为空，用于指定 retention.ms 等 topic 级配置。
func (a *AdminClient) CreateTopic(ctx context.Context, topic string, partitions, replicationFactor int, configEntries ...kafka.ConfigEntry) error {
	resp, err := a.client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
			ConfigEntries:     configEntries,
		}},
	})
	if err != nil {
		return fmt.Errorf("kafka admin client: create topic %q failed: %w", topic, err)
	}
	if e := resp.Errors[topic]; e != nil {
		return fmt.Errorf("kafka admin client: create topic %q failed: %w", topic, e)
	}
	return nil
}

// DeleteTopic 删除 topic。
func (a *AdminClient) DeleteTopic(ctx context.Context, topic string) error {
	resp, err := a.client.DeleteTopics(ctx, &kafka.DeleteTopicsRequest{Topics: []string{topic}})
	if err != nil {
		return fmt.Errorf("kafka admin client: delete topic %q failed: %w", topic, err)
	}
	if e := resp.Errors[topic]; e != nil {
		return fmt.Errorf("kafka admin client: delete topic %q failed: %w", topic, e)
	}
	return nil
}

// DeleteConsumerGroup 删除消费组。
func (a *AdminClient) DeleteConsumerGroup(ctx context.Context, group string) error {
	resp, err := a.client.DeleteGroups(ctx, &kafka.DeleteGroupsRequest{GroupIDs: []string{group}})
	if err != nil {
		return fmt.Errorf("kafka admin client: delete consumer group %q failed: %w", group, err)
	}
	if e := resp.Errors[group]; e != nil {
		return fmt.Errorf("kafka admin client: delete consumer group %q failed: %w", group, e)
	}
	return nil
}

// replicationFactor 取首个分区的副本数作为 topic 副本数（Kafka 要求 topic 内
// 所有分区副本数一致）。
func replicationFactor(t kafka.Topic) int {
	if len(t.Partitions) == 0 {
		return 0
	}
	return len(t.Partitions[0].Replicas)
}

// underReplicated 判断是否存在 ISR 数小于副本数的分区。
func underReplicated(t kafka.Topic) bool {
	for _, p := range t.Partitions {
		if len(p.Replicas) > 0 && len(p.Isr) < len(p.Replicas) {
			return true
		}
	}
	return false
}

func nilIfEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
