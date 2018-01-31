package controllers

import (
	"golang.org/x/net/websocket"
	"time"
	"github.com/golang/glog"
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego"

	"encoding/json"
	"io"
	"sync"
	"text/template"

	"dev-flows-api-golang/models/common"
	"dev-flows-api-golang/models/team2user"
	"dev-flows-api-golang/models/user"
	"dev-flows-api-golang/modules/pod"

	clustermodel "dev-flows-api-golang/models/cluster"
	sqlstatus "dev-flows-api-golang/models/sql/status"
	t2u "dev-flows-api-golang/models/team2user"
	clientmodule "dev-flows-api-golang/modules/client"
	streamClient "dev-flows-api-golang/modules/client"

	k8sWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/pkg/api/v1"
	v1beta1 "k8s.io/client-go/pkg/apis/batch/v1"

	"dev-flows-api-golang/models"
	"bytes"
	"net/http"
	"encoding/base64"
)

var (
	//webSocketWatchPod  watch pod
	webSocketWatchPod = make(map[string]map[string]WatchStruct)
	wsPodMux          sync.RWMutex

	//webSocketWatchJob watch Job
	webSocketWatchJob = make(map[string]map[string]WatchStruct)
	wsDepMux          sync.RWMutex

	//k8sDeploymentConnect watch job
	k8sDeploymentConnect = make(map[string]int)
	depConnMux           sync.RWMutex

	//k8sPodConnect watch pod
	k8sPodConnect = make(map[string]int)
	podConnMux    sync.RWMutex

	filterPodTime int64
)


//要用到map 用来存贮已经存在的job pod的信息 删除之后 也从map中删除 缓存机制 namespace
var BuildOfNamespace map[string]BuildMessage

type SocketLogController struct {
	BaseController
	resourceType string
	clusterID    string
	Teamspace    string
}

//AuthStruct auth
type AuthStruct struct {
	AccessToken string `json:"accessToken,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Teamspace   string `json:"teamspace,omitempty"`
	Type        string `json:"type,omitempty"`
	LogStruct   `json:",omitempty"`
	BuildMessage `json:",omitempty"`
}

//WatchStruct
type WatchStruct struct {
	Type      string
	Name      []string
	Cluster   string
	Conn      *websocket.Conn
	Namespace string `json:"namespace,omitempty"`
	Code      string
}

type LogStruct struct {
	Name          string
	Cluster       string
	Tail          int64    `json:"tail,omitempty"`
	ContainerName []string `json:"containerName, omitempty`
}

//PodMessage
type PodMessage struct {
	Type      string  `json:"type,omitempty"`
	Data      *v1.Pod `json:"data,omitempty"`
	WatchType string  `json:"watchType,omitempty"`
	Name      string  `json:"name,omitempty"`
}

//JobMessage
type JobMessage struct {
	Type      string              `json:"type,omitempty"`
	Data      *v1beta1.Job        `json:"data,omitempty"`
	WatchType string              `json:"watchType,omitempty"`
	Name      string              `json:"name,omitempty"`
}

type LogMessage struct {
	Name string
	LogTime time.Time
	Log  string `json:"log,omitempty"`
}

type ActionStruct struct {
	Action string `json:"action,omitempty"`
}

//Todo apiserver初始化的时候集群处于不可用状态，则标记。
//func init() {
//	go startInit()
//}

func startInit() {
	sql := common.NewDataSelectQuery(&common.PaginationQuery{From: 0, Size: 30}, nil, nil)
	list, _, err := clustermodel.GetClusterList(sql)
	if err != nil {
		beego.Error("Find cluster list error", err)
		return
	}
	s := &SocketLogController{}
	for _, value := range list.Clusters {
		err := connectk8sJob(s, value.ClusterID)
		if err == false {
			beego.Error("connect to " + value.ClusterID + " cluster Job error")
			continue
		}
		err = connectK8sPod(s, value.ClusterID)
		if err == false {
			beego.Error("connect to " + value.ClusterID + " cluster pod error")
			continue
		}
	}
}

func connectk8sJob(s *SocketLogController, clusterID string) bool {
	//filterDeploymentTime = time.Now().Unix()
	client, isOk := s.GetClientSetOrRespErr(clusterID)
	if isOk == false {
		return false
	}
	api := streamClient.GetStreamApi(client, "")
	result, err := api.WatchResource("job")
	if err != nil {
		return false
	}
	depConnMux.Lock()
	k8sDeploymentConnect[clusterID] = 1
	depConnMux.Unlock()
	go startWatchJob(clusterID, &result)
	podConnMux.Lock()
	k8sPodConnect[clusterID] = 1
	podConnMux.Unlock()
	return true
}

func connectK8sPod(s *SocketLogController, clusterID string) bool {
	filterPodTime = time.Now().Unix()
	client, isOk := s.GetClientSetOrRespErr(clusterID)
	if isOk == false {
		return false
	}
	api := streamClient.GetStreamApi(client, "")
	result, err := api.WatchResource("pod")
	if err != nil {
		podConnMux.Lock()
		k8sPodConnect[clusterID] = 0
		podConnMux.Unlock()
		return false
	}
	go startWatchPod(clusterID, &result)
	podConnMux.Lock()
	k8sPodConnect[clusterID] = 1
	podConnMux.Unlock()
	return true
}

func startWatchJob(clusterID string, result *k8sWatch.Interface) {
	readOnlyChan := (*result).ResultChan()
	defer connectk8sJob(&SocketLogController{}, clusterID)
	for {
		select {
		case event, isOpen := <-readOnlyChan:
			if isOpen == false {
				return
			}
			// k8s data error
			wsDepMux.RLock()
			watcher := webSocketWatchJob[clusterID]
			wsDepMux.RUnlock()
			if len(watcher) == 0 {
				continue
			}
			var content []byte
			var parseErr error
			dm, parseIsOk := event.Object.(*v1beta1.Job)
			if false == parseIsOk {
				return
			}
			var message JobMessage
			for key, value := range watcher {
				if dm.ObjectMeta.Namespace != value.Namespace {
					continue
				}
				for _, name := range value.Name {
					if value.Type == "app" {
						if dm.ObjectMeta.Labels["tenxcloud.com/appName"] != name {
							continue
						}
						message = JobMessage{
							Type:      string(event.Type),
							Data:      dm,
							WatchType: value.Type,
							Name:      name,
						}
					}
					if value.Type == "job" {
						if dm.ObjectMeta.Labels["name"] != name {
							continue
						}
						message = JobMessage{
							Type:      string(event.Type),
							Data:      dm,
							WatchType: value.Type,
							Name:      name,
						}
					}
					content, parseErr = json.Marshal(message)
					if parseErr != nil {
						beego.Error(clusterID, "parseErr deployment to brower failed", parseErr)
						value.Conn.Close()
						delete(watcher, key)
						continue
					}
					nerr := websocket.Message.Send(value.Conn, string(content[:]))
					if nerr != nil {
						value.Conn.Close()
						delete(watcher, key)
					}
				}
			}
		}
	}
}

func startWatchPod(clusterID string, result *k8sWatch.Interface) {
	readOnlyChan := (*result).ResultChan()
	defer func() {
		connectK8sPod(&SocketLogController{}, clusterID)
	}()
	for {
		select {
		case event, isOpen := <-readOnlyChan:
			// k8s data error
			if isOpen == false {
				depConnMux.Lock()
				k8sDeploymentConnect[clusterID] = 0
				depConnMux.Unlock()
				return
			}
			wsPodMux.RLock()
			watcher := webSocketWatchPod[clusterID]
			wsPodMux.RUnlock()
			if len(watcher) == 0 {
				continue
			}
			var content []byte
			var parseErr error
			pod, parseIsOk := event.Object.(*v1.Pod)
			if false == parseIsOk {
				depConnMux.Lock()
				k8sDeploymentConnect[clusterID] = 0
				depConnMux.Unlock()
				//重新建立连接
				return
			}
			if pod.Status.StartTime != nil && filterPodTime > pod.Status.StartTime.Unix() {
				continue
			}
			var message PodMessage
			for key, value := range watcher {
				if pod.ObjectMeta.Namespace != value.Namespace {
					continue
				}
				for _, name := range value.Name {
					if pod.ObjectMeta.Labels["name"] != name {
						continue
					}
					message = PodMessage{
						Type:      string(event.Type),
						Data:      pod,
						WatchType: value.Type,
						Name:      name,
					}
					content, parseErr = json.Marshal(message)
					if parseErr != nil {
						value.Conn.Close()
						beego.Error(clusterID, "parseErr job to brower failed", parseErr)
						delete(watcher, key)
						continue
					}
					nerr := websocket.Message.Send(value.Conn, string(content[:]))
					if nerr != nil {
						value.Conn.Close()
						delete(watcher, key)
					}
				}
			}
		}
	}
}

//CreateAccessToken
// @Success 200 success
// @router /token [get]
func (s *SocketLogController) CreateAccessToken() {
	userToken := s.User.APIToken
	e := md5.Sum([]byte(userToken))
	s.ResponseSuccess(fmt.Sprintf("%x", e))
	return
}

//CreateWebSocketConn is used created websocket connection
// @Success 200 success
// @router / [get]
func (s *SocketLogController) CreateWebSocketConn() {
	server := websocket.Server{
		Handler: s.handleWebSocket,
	}
	server.ServeHTTP(s.Ctx.ResponseWriter, s.Ctx.Request)
}

func (s *SocketLogController) handleWebSocket(conn *websocket.Conn) {
	var auth string
	conn.SetDeadline(time.Now().Add(time.Second * 1000))
	websocket.Message.Receive(conn, &auth)
	var authStruct AuthStruct
	e := json.Unmarshal([]byte(auth), &authStruct)
	if e != nil {
		websocket.Message.Send(conn, `{"status": false}`)
		s.CloseWebSocket(conn)
		return
	}
	if authStruct.AccessToken == "" || authStruct.Namespace == "" {
		websocket.Message.Send(conn, `{"status": false}`)
		s.CloseWebSocket(conn)
		return
	}
	isauth := checkTokenDemo(authStruct.AccessToken, authStruct.Namespace, authStruct.Teamspace)
	if isauth == false {
		websocket.Message.Send(conn, `{"status": false}`)
		s.CloseWebSocket(conn)
		return
	}
	//检查校验参数
	//if message := CheckLogData(authStruct); message != "" {
	//	websocket.Message.Send(conn, message)
	//	s.CloseWebSocket(conn)
	//	return
	//}

	websocket.Message.Send(conn, `{"status": true}`)
	s.Namespace = authStruct.Namespace
	s.Teamspace = authStruct.Teamspace
	channel := make(chan int)
	if authStruct.Type == "log" {
		logStruct := authStruct.LogStruct
		go handleLogWebsocket(conn, s, channel, &logStruct)
	} else {
		go handleStatusWebsocket(conn, s, channel)
	}
	<-channel
	return
}



func (s *SocketLogController) GetStageBuildLogsFromK8S(authStruct AuthStruct,conn *websocket.Conn){

	method := "SocketLogController.GetStageBuildLogsFromK8S"

	imageBuilder:=models.NewImageBuilder(s.clusterID)

	build,err:=s.GetValidStageBuild(authStruct)
	if err!=nil{
		glog.Errorf("%s get log from k8s failed:===>>%v\n",method,err)
		websocket.Message.Send(conn,err)
		s.CloseWebSocket(conn)
		return
	}

	if build.Status==common.STATUS_WAITING{
		buildStatus:= struct {
			BuildStatus string `json:"buildStatus"`
		}{
			BuildStatus:"waiting",
		}

		websocket.Message.Send(conn,buildStatus)
		s.CloseWebSocket(conn)
		return
	}

	if build.PodName==""{
		podName,err:=imageBuilder.GetPodName(build.Namespace,build.JobName)
		if err!=nil||podName==""{
			glog.Errorf("%s get job name=[%s] pod name failed:======>%v\n",method,build.JobName,err)
			websocket.Message.Send(conn,err)
			s.CloseWebSocket(conn)
			return
		}

		models.NewCiStageBuildLogs().UpdatePodNameById(podName,build.BuildId)

		websocket.Message.Send(conn,"")

		return
	}


	return
}


func (s *SocketLogController) GetLogsFromK8S(authStruct AuthStruct,imageBuilder models.ImageBuilder,conn *websocket.Conn){
	method := "SocketLogController.GetStageBuildLogsFromK8S"
	build,err:=s.GetValidStageBuild(authStruct)
	if err!=nil{
		glog.Errorf("%s get log from k8s failed:===>>%v\n",method,err)
		websocket.Message.Send(conn,err)
		s.CloseWebSocket(conn)
		return
	}

	if build.Status==common.STATUS_WAITING{
		buildStatus:= struct {
			BuildStatus string `json:"buildStatus"`
		}{
			BuildStatus:"waiting",
		}

		websocket.Message.Send(conn,buildStatus)
		s.CloseWebSocket(conn)
		return
	}

	if build.PodName==""{
		podName,err:=imageBuilder.GetPodName(build.Namespace,build.JobName)
		if err!=nil||podName==""{
			glog.Errorf("%s get job name=[%s] pod name failed:======>%v\n",method,build.JobName,err)
			websocket.Message.Send(conn,err)
			s.CloseWebSocket(conn)
			return
		}

		models.NewCiStageBuildLogs().UpdatePodNameById(podName,build.BuildId)
		//return _getLogsFromK8S(imageBuilder, build.namespace, build.job_name, podName, socket)
		websocket.Message.Send(conn,"")

		return
	}


	return
}




func (s *SocketLogController) GetValidStageBuild(authStruct AuthStruct) (models.CiStageBuildLogs, error) {
	var build models.CiStageBuildLogs
	method := "SocketLogController.GetValidStageBuild"
	stage, err := models.NewCiStage().FindOneById(authStruct.StageId)
	if err != nil {
		glog.Errorf("%s find stage by stageId failed or not exist from database: %v\n", method, err)
		return build, err
	}
	if stage.FlowId != authStruct.FlowId {

		return build, fmt.Errorf("Stage is not %s in the flow", authStruct.StageId)
	}

	build, err = models.NewCiStageBuildLogs().FindOneById(authStruct.StageBuildId)
	if err != nil {
		glog.Errorf("%s find stagebuild by StageBuildId failed or not exist from database: %v\n", method, err)
		return build, err
	}

	if stage.StageId != build.StageId {

		return build, fmt.Errorf("Build is not %s one of the stage", build.BuildId)

	}

	return build, nil
}



func (s *SocketLogController) CloseWebSocket(conn *websocket.Conn) {
	conn.Close()
	return
}

func  CheckLogData(authStruct EnnFlow) string {
	method := "CheckLogData"
	if authStruct.FlowId == "" || authStruct.StageId == "" || authStruct.StageBuildId == "" {
		glog.Errorf("%s Missing parameters \n", method)
		return `<font color="red">[Enn Flow API Error] Missing parameters.</font>`
	}

	return ""
}

func handleStatusWebsocket(conn *websocket.Conn, s *SocketLogController, channel chan int) {
	method := "receiveMessage"
	var name string
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf(fmt.Sprintf("%s recovered form panic %v", method, r))
		}
		conn.Close()
		channel <- 1
	}()
	for {
		err := websocket.Message.Receive(conn, &name)
		if nil != err {
			beego.Warn(method, "Error event from user client:", err)
			return
		}
		if name == "0" {
			websocket.Message.Send(conn, `{"code": 1}`)
			continue
		}

		var watchStruct WatchStruct
		isOk := json.Unmarshal([]byte(name), &watchStruct)
		if isOk != nil {
			beego.Error(method, "Parse json err", isOk)
			return
		}
		watchStruct.Namespace = s.Namespace
		watchStruct.Conn = conn
		key := fmt.Sprintf("%x", &conn)
		if len(watchStruct.Cluster) > 0 && len(watchStruct.Namespace) > 0 {
			depConnMux.Lock()
			if k8sDeploymentConnect[watchStruct.Cluster] == 0 {
				k8sDeploymentConnect[watchStruct.Cluster] = 0
			}
			depConnMux.Unlock()
			podConnMux.Lock()
			if k8sPodConnect[watchStruct.Cluster] == 0 {
				k8sPodConnect[watchStruct.Cluster] = 0
			}
			podConnMux.Unlock()
			if watchStruct.Type == "pod" {
				wsPodMux.Lock()
				_, isExists := webSocketWatchPod[watchStruct.Cluster]
				if isExists == false {
					webSocketWatchPod[watchStruct.Cluster] = make(map[string]WatchStruct)
				}
				webSocketWatchPod[watchStruct.Cluster][key] = watchStruct
				wsPodMux.Unlock()
			} else {
				wsDepMux.Lock()
				_, isExists := webSocketWatchJob[watchStruct.Cluster]
				if isExists == false {
					webSocketWatchJob[watchStruct.Cluster] = make(map[string]WatchStruct)
				}
				webSocketWatchJob[watchStruct.Cluster][key] = watchStruct
				wsDepMux.Unlock()
			}
			continue
		}
		if watchStruct.Code == "CLOSED" {
			if watchStruct.Type == "pod" {
				wsPodMux.Lock()
				delete(webSocketWatchPod[watchStruct.Cluster], key)
				wsPodMux.Unlock()
			} else {
				wsDepMux.Lock()
				delete(webSocketWatchJob[watchStruct.Cluster], key)
				wsDepMux.Unlock()
			}
			continue
		}
		return
	}
}

func handleLogWebsocket(conn *websocket.Conn, s *SocketLogController, channel chan int, logStruct *LogStruct) {
	method := "handleLogWebsocket"
	haveErr := false
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf(fmt.Sprintf("%s recovered form panic %v", method, r))
		}
		if true == haveErr {
			conn.Close()
			channel <- 1
		}
	}()
	client, isOk := s.GetClientSetOrRespErr(logStruct.Cluster)
	if false == isOk {
		haveErr = true
		beego.Error(method, "Get cluster client err", logStruct.Cluster)
		return
	}
	namespace := s.Namespace
	if "" != s.Teamspace {
		namespace = s.Teamspace
	}
	logApi := streamClient.GetStreamApi(client, namespace)
	var containerName []string
	if nil != logStruct.ContainerName {
		containerName = logStruct.ContainerName
	} else {
		podDetail, err := pod.GetPodDetail(*client.CoreV1Client, namespace, logStruct.Name)
		if nil != err {
			haveErr = true
			beego.Error(method, "Get the pod which name is "+logStruct.Name, err)
			return
		}
		containerName = make([]string, len(podDetail.Containers))
		for index, container := range podDetail.Containers {
			containerName[index] = container.Name
		}
	}
	for _, container := range containerName {
		if "" == container {
			container = logStruct.Name
		}
		reader, err := logApi.FollowLog(logStruct.Name, container, logStruct.Tail)
		if nil != err {
			haveErr = true
			beego.Error(method, "Follow log failed, Pod nam is ", logStruct.Name, err)
			return
		}
		go sendLogToBrower(conn, container, reader, channel)
	}
}

//sendLogToBrower recevie from k8s and send log to brower
func sendLogToBrower(conn *websocket.Conn, containerName string, reader io.ReadCloser, channel chan int) {
	defer func() {
		if r := recover(); r != nil {

			glog.Errorf(fmt.Sprintf("sendLogToBrower recovered form panic %v", r))
		}
		conn.Close()
		reader.Close()
		channel <- 1
	}()
	data := make([]byte, 1024*1024, 1024*1024)
	isPause := make(chan int, 1)
	isContinue := make(chan int, 1)
	go pauseLog(isPause, isContinue, conn)
	for {
		n, err := reader.Read(data)
		if nil != err {
			if err == io.EOF {
				sendLog([]byte("TENXCLOUD_END_OF_STREAM"), conn)
				return
			}
			beego.Error("Read container log failed, containername is", containerName, ", err is ", err)
			return
		}
		logMessage := &LogMessage{
			Name: containerName,
			Log:  template.HTMLEscapeString(string(data[:n])),
		}
		message, err := json.Marshal(logMessage)
		if nil != err {
			beego.Error("Parse container log failed, container name is ", containerName, ", err is ", err)
			return
		}
		select {
		case <-isPause:
			<-isContinue
			err := sendLog(message, conn)
			if nil != err {
				beego.Error("Send log err, container name is ", containerName, " err is ", err)
				return
			}
		default:
			err := sendLog(message, conn)
			if nil != err {
				beego.Error("Send log err, container name is ", containerName, " err is ", err)
				return
			}
		}
	}
}

//sendLog Send log to brower
func sendLog(data []byte, conn *websocket.Conn) error {
	err := websocket.Message.Send(conn, string(data[:]))
	return err
}

//pauseLog pause and continue send log
func pauseLog(isPause chan int, isContinue chan int, conn *websocket.Conn) {
	defer conn.Close()
	for {
		var actionStruct ActionStruct
		var action string
		err := websocket.Message.Receive(conn, &action)
		json.Unmarshal([]byte(action), &actionStruct)
		if nil != err {
			if err == io.EOF {
				return
			}
			beego.Error("Receive pause error ", err)
			return
		}
		if "pause" == actionStruct.Action {
			isPause <- 1
			err = websocket.Message.Receive(conn, &action)
			if nil != err {
				beego.Error("Receive play error ", err)
				return
			}
			json.Unmarshal([]byte(action), &actionStruct)
			if "play" == actionStruct.Action {
				isContinue <- 1
			}
			continue
		}
		continue
	}
}

// get cluster configuration from database
// automatically response to client if error happens
func (w *SocketLogController) GetClusterOrRespErr(clusterId string) (*clustermodel.ClusterModel, bool) {
	cluster := &clustermodel.ClusterModel{}
	if statusCode, err := cluster.Get(clusterId); statusCode != sqlstatus.SQLSuccess {
		if statusCode == sqlstatus.SQLErrNoRowFound {
			glog.Errorf("Cluster %s does not exist\n", clusterId)
			return nil, false
		}
		glog.Errorf("Failed to query cluster from db, error:%s\n", err)
		return nil, false
	}
	return cluster, true
}

func (w *SocketLogController) GetClientSetOrRespErr(clusterId string) (*clientmodule.ClientSet, bool) {
	// 获取cluster配置
	cluster, ok := w.GetClusterOrRespErr(clusterId)
	if false == ok {
		glog.Errorf("Failed to get cluster configuration: %s\n", clusterId)
		return nil, false
	}

	//初始化client
	cs, err := clientmodule.NewClientSet(cluster.ClusterName, cluster.APIProtocol, cluster.APIHost, cluster.APIToken, cluster.APIVersion)
	if nil != err {
		glog.Errorf("Failed to init k8s api of %s, error:%s\n", clusterId, err)
		return nil, false
	}
	return cs, true
}

//TODO 检测当前namespace是否可以操作传入teamspace
func checkTokenDemo(token, namespace, teamspace string) bool {
	userModel := &user.UserModel{}
	_, err := userModel.GetByNamespace(namespace)
	if err != nil {
		return false
	}
	if userModel.Role != team2user.SuperAdminUser {
		count := 1
		if "" != teamspace {
			t := t2u.NewTeamUserModel()
			count, err = t.AuthUserCanUseTeamspace(userModel.UserID, teamspace)
		}
		if 0 == count || nil != err {
			return false
		}

	}
	userToken := md5.Sum([]byte(userModel.APIToken))
	return token == fmt.Sprintf("%x", userToken)
}

func newId(r *http.Request) string {

	hash := fmt.Sprintf("%s %s", r.RemoteAddr, time.Now())
	buf := bytes.NewBuffer(nil)
	sum := md5.Sum([]byte(hash))
	encoder := base64.NewEncoder(base64.URLEncoding, buf)
	encoder.Write(sum[:])
	encoder.Close()
	return buf.String()[:20]
}