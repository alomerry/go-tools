package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

func (c *Client) CreateResourceByYaml(yamlContent string) error {
	// 分割 YAML 为多个文档
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(yamlContent)), 4096)
	for {
		// 解码为 Unstructured 对象
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if len(rawObj.Raw) == 0 {
			// 跳过空文档
			continue
		}

		// 使用反序列化器解码
		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			logrus.Errorf("解码失败: %v", err)
			return err
		}

		// 转换为 Unstructured 类型
		unstructuredObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			logrus.Errorf("无法将对象转换为 Unstructured")
			return fmt.Errorf("无法将对象转换为 Unstructured")
		}

		// 获取资源对应的 GVR
		mapping, err := c.GetMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			logrus.Errorf("获取资源对应的 GVR 失败: %v", err)
			return err
		}

		// 确定命名空间
		namespace := unstructuredObj.GetNamespace()
		if namespace == "" {
			namespace = "default" // 默认命名空间
		}

		// 创建资源
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			dr = c.GetDynamicClient().Resource(mapping.Resource).Namespace(namespace)
		} else {
			dr = c.GetDynamicClient().Resource(mapping.Resource)
		}

		createdObj, err := dr.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			logrus.Errorf("创建资源失败: %v", err)
			return err
		}

		logrus.Infof("资源创建成功: %s/%s\n", createdObj.GetKind(), createdObj.GetName())
	}
	return nil
}
