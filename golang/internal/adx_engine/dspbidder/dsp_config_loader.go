package dspbidder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ConfigLoader DSP配置加载器
type ConfigLoader struct {
	client     *minio.Client
	bucketName string
	prefix     string
	ctx        context.Context
}

// S3Config S3配置
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Prefix          string
	UseSSL          bool
	ScanInterval    time.Duration
	Timeout         time.Duration
}

// NewConfigLoader 创建DSP配置加载器
func NewConfigLoader(config *S3Config) (*ConfigLoader, error) {
	ctx := context.Background()

	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return &ConfigLoader{
		client:     client,
		bucketName: config.BucketName,
		prefix:     config.Prefix,
		ctx:        ctx,
	}, nil
}

// DSPInfo DSP配置信息
type DSPInfo struct {
	DSPID       string                 `json:"dsp_id"`
	DSPName     string                 `json:"dsp_name"`
	Status      string                 `json:"status"`
	QPSLimit    int                    `json:"qps_limit"`
	BudgetDaily float64                `json:"budget_daily"`
	Endpoint    string                 `json:"endpoint"`
	AuthToken   string                 `json:"auth_token"`
	Timeout     time.Duration          `json:"timeout"`
	RetryCount  int                    `json:"retry_count"`
	RetryDelay  time.Duration          `json:"retry_delay"`
	Targeting   *DSPTargeting          `json:"targeting,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     string                 `json:"version,omitempty"`
}

// DSPTargeting DSP定向配置
type DSPTargeting struct {
	IndexingDoc []IndexingClause `json:"indexingdoc"`
}

// IndexingClause 索引条款
type IndexingClause struct {
	ClauseID    string      `json:"clause_id"`
	Description string      `json:"description"`
	Conditions  []Condition `json:"conditions"`
}

// Condition 定向条件
type Condition struct {
	Field    string   `json:"field"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// ListDSPFiles 列出所有DSP配置文件
func (s *ConfigLoader) ListDSPFiles() ([]string, error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	var dspFiles []string

	for object := range s.client.ListObjects(s.ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    s.prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("list objects error: %w", object.Err)
		}

		// 只处理 .json 文件
		if strings.HasSuffix(object.Key, ".json") {
			dspFiles = append(dspFiles, object.Key)
		}
	}

	// 按文件路径排序，确保一致性
	sort.Strings(dspFiles)
	return dspFiles, nil
}

// ReadDSPFile 读取单个DSP配置文件
func (s *ConfigLoader) ReadDSPFile(objectKey string) (*DSPInfo, error) {
	obj, err := s.client.GetObject(s.ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", objectKey, err)
	}
	defer obj.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, fmt.Errorf("failed to read object %s: %w", objectKey, err)
	}

	dspInfo := &DSPInfo{}
	if err := json.Unmarshal(buf.Bytes(), dspInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DSP info from %s: %w", objectKey, err)
	}

	dspInfo.UpdatedAt = time.Now()
	return dspInfo, nil
}

// ReadAllDSPs 读取所有DSP配置
func (s *ConfigLoader) ReadAllDSPs() (map[string]*DSPInfo, error) {
	objectKeys, err := s.ListDSPFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list DSP files: %w", err)
	}

	dspMap := make(map[string]*DSPInfo)
	for _, objectKey := range objectKeys {
		dspInfo, err := s.ReadDSPFile(objectKey)
		if err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("failed to read DSP file %s: %v\n", objectKey, err)
			continue
		}

		if dspInfo.DSPID == "" {
			return nil, fmt.Errorf("DSP ID is empty in file %s", objectKey)
		}

		dspMap[dspInfo.DSPID] = dspInfo
	}

	return dspMap, nil
}

// WatchDSPChanges 监控DSP配置变化
func (s *ConfigLoader) WatchDSPChanges() <-chan map[string]*DSPInfo {
	changeCh := make(chan map[string]*DSPInfo, 1)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		// 初始加载
		dspMap, err := s.ReadAllDSPs()
		if err == nil {
			select {
			case changeCh <- dspMap:
			default:
			}
		}

		for {
			select {
			case <-ticker.C:
				newDspMap, err := s.ReadAllDSPs()
				if err == nil {
					select {
					case changeCh <- newDspMap:
					default:
					}
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return changeCh
}
