package main

import (
	"app/table/coreClass/centControlPlatform"
	"app/table/coreClass/middleAgentRtc"
	"app/table/coreClass/publicStruct"
	"encoding/base64"
	"encoding/json"

	"cnb.cool/accbot/goTool/sessionPkg"
	"github.com/ghp3000/logs"
)

var s *sessionPkg.Session

func main() {

	_ = centControlPlatform.NewCore("minio.accjs.cn", "10003983fcece0c60311f7a738057eb2260181769675270297")
	//Server.Login()
	////
	////从集控中心更新设备列表
	//deviceList, err := Server.CenterGetDeviceList()
	//if err != nil {
	//	logs.Error("从集控中心获取设备列表失败", err)
	//	return
	//}
	//
	//logs.Info("测试从集控中心获取设备列表成功,list=%V", deviceList)
	//
	////
	//////////从缓存中获取所有设备
	//devices := Server.GetAllDevice()
	//for _, device := range devices {
	//	logs.Info("测试从缓存中获取所有设备，设备ID:%d", device.DeviceId)
	//}
	//time.Sleep(time.Second * 5)
	//////发起一个中间件RTC，并创建对应中间件的对象
	////_ = 测试连接中间件(Server)
	////time.Sleep(time.Second * 5)
	//middleAgent, err := Server.CreatMiddlewareRtc(97, nil, nil, nil)
	//if err != nil {
	//	logs.Error("创建中间件RTC失败", err)
	//	return
	//}
	//logs.Info("测试创建中间件RTC成功,设备ID:%d", middleAgent.MiddlewareId)
	//agent.SyncGetDeviceDetails(24837)
	//agent.AsyncGetAllDeviceDetails([]uint64{24837})
	//time.Sleep(time.Second * 5)
	//go 测试创建端口映射打洞(Server, middleAgent.Devices[1].DeviceId, 9999, 9999, 0)
	//middleAgent.SyncDelayedetection()
	//go 同步功能测试(Server, middleAgent)
	//rtc, err := Server.CreatDeviceH264AudioRtc(devices[1].DeviceId, nil, nil, nil)
	//if err != nil {
	//	return
	//}
	//time.Sleep(time.Second * 5)
	//rtc.AsyncStartVideo(1)
	//rtc.AsyncStartAudio(1)
	//
	//time.Sleep(time.Second * 5)
	//rtc.AsyncStopVideo()
	//rtc.AsyncStopAudio()

	select {}
}

// 测试连接接中间件
func 测试连接中间件(s *centControlPlatform.Core) *middleAgentRtc.TypeInfo {
	middleWareIds := s.GetAllMiddleWareIds()
	logs.Info("所有中间件ID:", middleWareIds)
	middleAgent, err := s.CreatMiddlewareRtc(middleWareIds[0], nil, nil, nil)
	if err != nil {
		logs.Error("创建中间件RTC失败", err)
		return nil
	}
	return middleAgent
}

// 同步功能测试
func 同步功能测试(server *centControlPlatform.Core, middleAgent *middleAgentRtc.TypeInfo) {
	//获取设备图片测试
	测试获取图片(middleAgent)

	//测试创建端口映射打洞

}

func 测试获取图片(middleAgent *middleAgentRtc.TypeInfo) {
	deviceId := middleAgent.Devices[1].DeviceId
	base64String, err := middleAgent.SyncGetDeviceImageToBase64(deviceId, 1080, 2)
	if err != nil {
		return
	}
	var picString publicStruct.Base64Image
	err = json.Unmarshal([]byte(base64String), &picString)
	if err != nil {
		logs.Error("json反序列化失败", err)
		return
	}

	// 解码base64字符串
	decodedBytes, err := base64.StdEncoding.DecodeString(picString.Base64Img)
	if err != nil {
		logs.Error("base64解码失败", err)
		return
	}
	middleAgentRtc.WriteBytesToFile(decodedBytes, "D:\\Users\\Desktop\\yunshouji\\ApiAgent\\data\\deviceImage\\a.webp")
	logs.Info("测试获取图片成功，设备ID:%d，图片保存路径:%s", deviceId, "D:\\Users\\Desktop\\yunshouji\\ApiAgent\\data\\deviceImage\\a.webp")
}
func 测试创建端口映射打洞(server *centControlPlatform.Core, deviceId uint64, phonePort int, localPort int, mode int) {
	// 创建端口映射打洞对象
	_, err := server.CreatPortMapSocket5Rtc(deviceId, phonePort, localPort, mode)
	if err != nil {
		logs.Error("创建端口映射[%d]对象失败%v", deviceId, err)
		return
	}

}
