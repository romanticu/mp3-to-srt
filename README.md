# mp3-to-srt
mp3 file to srt file

## use case
### mp3-trans/config.go 改成自己的配置
```go
const (
	// ===== 阿里云Oss对象服务配置 =====
	// OSS 对外服务的访问域名
	AliyunOss_Endpoint = "oss-cn-beijing.aliyuncs.com"
	// 存储空间（Bucket）名称
	AliyunOss_BucketName = "your oss bucket name"
	// 存储空间（Bucket 域名）地址
	AliyunOss_BucketDomain = "your oss bucket domain"
	AliyunOss_AccessKeyId = "your oss access key id"
	AliyunOss_AccessKeySecret = "your oss access key secret"
	
	
	
	
	// ===== 阿里云语音识别配置 =====
	AliyunClound_AccessKeyId = "your aliyun clound access key id"
	AliyunClound_AccessKeySecret = "your aliyun clound access key secret"
	AliyunClound_AppKey = "your aliyun clound app key"
)
```
### 运行程序
```
// go run main.go filePath
go run main.go test.wav
```
