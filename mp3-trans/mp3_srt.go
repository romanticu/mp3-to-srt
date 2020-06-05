package mp3_trans

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/dykily/mp3srt/mp3-trans/ali_yun"
	"os"
	"path"
	"strconv"
)

type MP3toSrt struct {
	AliyunOss        ali_yun.AliyunOss   //oss
	AliyunClound     ali_yun.AliyunCloud //语音识别引擎
	IntelligentBlock bool                //智能分段处理
	TempDir          string              //临时文件目录
	//AppDir           string               //应用根目录
}


// 获取应用
func NewApp() *MP3toSrt {
	app := ReadConfig()

	return app
}


// 配置
func ReadConfig () *MP3toSrt {
	appconfig := &MP3toSrt{}

	//AliyunOss
	appconfig.AliyunOss.Endpoint = AliyunOss_Endpoint
	appconfig.AliyunOss.AccessKeyId = AliyunOss_AccessKeyId
	appconfig.AliyunOss.AccessKeySecret = AliyunOss_AccessKeySecret
	appconfig.AliyunOss.BucketName = AliyunOss_BucketName
	appconfig.AliyunOss.BucketDomain = AliyunOss_BucketDomain

	//AliyunCloud
	appconfig.AliyunClound.AccessKeyId = AliyunCloud_AccessKeyId
	appconfig.AliyunClound.AccessKeySecret = AliyunCloud_AccessKeySecret
	appconfig.AliyunClound.AppKey = AliyunCloud_AppKey


	appconfig.IntelligentBlock = true
	appconfig.TempDir = "temp/audio"

	return appconfig
}


// 应用初始化
func (app *MP3toSrt) Init(appDir string) {
	//app.AppDir = appDir
}

// handler mp3
func (app *MP3toSrt) RunMP3(mp3 string)  {
	if mp3 == "" {
		panic("enter a video file waiting to be processed .")
	}

	//上传音频至OSS
	filelink := UploadAudioToCloud(app.AliyunOss , mp3)
	//获取完整链接
	filelink = "https://" + app.AliyunOss.GetObjectFileUrl(filelink)

	Log(filelink)
	Log("上传文件成功 , 识别中 ...")

	//阿里云录音文件识别
	AudioResult := AliyunAudioRecognition(app.AliyunClound, filelink , app.IntelligentBlock)

	Log("文件识别成功 , 字幕处理中 ...")

	//输出字幕文件
	AliyunAudioResultMakeSubtitleFile(mp3 , AudioResult)

}




//上传音频至oss
func UploadAudioToCloud(target ali_yun.AliyunOss , audioFile string) string {
	name := ""
	//提取文件名称
	if fileInfo, e := os.Stat(audioFile);e != nil {
		panic(e)
	} else {
		name = fileInfo.Name()
	}

	//上传
	if file , e := target.UploadFile(audioFile , name); e != nil {
		panic(e)
	} else {
		return file
	}
}


//阿里云录音文件识别
func AliyunAudioRecognition(engine ali_yun.AliyunCloud, filelink string , intelligent_block bool) (AudioResult map[int64][] *ali_yun.AliyunAudioRecognitionResult) {
	//创建识别请求
	taskid, client, e := engine.NewAudioFile(filelink)
	if e != nil {
		panic(e)
	}

	AudioResult = make(map[int64][] *ali_yun.AliyunAudioRecognitionResult)

	//遍历获取识别结果
	engine.GetAudioFileResult(taskid , client , func(result []byte) {
		fmt.Println("-----------")
		fmt.Println(string( result ))
		fmt.Println("-----------")

		//结果处理
		statusText, _ := jsonparser.GetString(result, "StatusText") //结果状态
		if statusText == ali_yun.STATUS_SUCCESS {

			//智能分段
			if intelligent_block {
				ali_yun.AliyunAudioResultWordHandle(result , func(vresult *ali_yun.AliyunAudioRecognitionResult) {
					channelId := vresult.ChannelId

					_ , isPresent  := AudioResult[channelId]
					if isPresent {
						//追加
						AudioResult[channelId] = append(AudioResult[channelId] , vresult)
					} else {
						//初始
						AudioResult[channelId] = []*ali_yun.AliyunAudioRecognitionResult{}
						AudioResult[channelId] = append(AudioResult[channelId] , vresult)
					}
				})
				return
			}

			_, err := jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				text , _ := jsonparser.GetString(value, "Text")
				channelId , _ := jsonparser.GetInt(value, "ChannelId")
				beginTime , _ := jsonparser.GetInt(value, "BeginTime")
				endTime , _ := jsonparser.GetInt(value, "EndTime")
				silenceDuration , _ := jsonparser.GetInt(value, "SilenceDuration")
				speechRate , _ := jsonparser.GetInt(value, "SpeechRate")
				emotionValue , _ := jsonparser.GetInt(value, "EmotionValue")

				vresult := &ali_yun.AliyunAudioRecognitionResult {
					Text:text,
					ChannelId:channelId,
					BeginTime:beginTime,
					EndTime:endTime,
					SilenceDuration:silenceDuration,
					SpeechRate:speechRate,
					EmotionValue:emotionValue,
				}

				_ , isPresent  := AudioResult[channelId]
				if isPresent {
					//追加
					AudioResult[channelId] = append(AudioResult[channelId] , vresult)
				} else {
					//初始
					AudioResult[channelId] = []*ali_yun.AliyunAudioRecognitionResult{}
					AudioResult[channelId] = append(AudioResult[channelId] , vresult)
				}
			} , "Result", "Sentences")
			if err != nil {
				panic(err)
			}
		}
	})

	return
}


//阿里云录音识别结果集生成字幕文件
func AliyunAudioResultMakeSubtitleFile(video string , AudioResult map[int64][] *ali_yun.AliyunAudioRecognitionResult)  {
	subfileDir := path.Dir(video)
	subfile := GetFileBaseName(video)
	Log(subfileDir, subfile)
	Log(len(AudioResult))
	for channel,result := range AudioResult {
		thisfile := subfileDir + "/" + subfile + "_channel_" +  strconv.FormatInt(channel , 10) + ".srt"
		//输出字幕文件
		Log(thisfile)
		file, e := os.Create(thisfile)
		if e != nil {
			panic(e)
		}

		defer file.Close() //defer

		index := 0
		for _ , data := range result {
			linestr := MakeSubtitleText(index , data.BeginTime , data.EndTime , data.Text)

			file.WriteString(linestr)

			index++
		}
	}
}


//拼接字幕字符串
func MakeSubtitleText(index int , startTime int64 , endTime int64 , text string) string {
	var content bytes.Buffer
	content.WriteString(strconv.Itoa(index))
	content.WriteString("\n")
	content.WriteString(SubtitleTimeMillisecond(startTime))
	content.WriteString(" --> ")
	content.WriteString(SubtitleTimeMillisecond(endTime))
	content.WriteString("\n")
	content.WriteString(text)
	content.WriteString("\n")
	content.WriteString("\n")
	return content.String()
}