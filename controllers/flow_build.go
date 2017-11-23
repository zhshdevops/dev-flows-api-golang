package controllers

import (
	"github.com/golang/glog"
	"fmt"
	"dev-flows-api-golang/models/user"
	"dev-flows-api-golang/models/common"
	//apiv1 "k8s.io/client-go/pkg/api/v1"
	"time"
	"k8s.io/client-go/1.4/pkg/apis/batch/v1"
	apiv1 "k8s.io/client-go/1.4/pkg/api/v1"
	"dev-flows-api-golang/models"
)

func MakeScriptEntryEnvForInitContainer(user *user.UserModel, containerInfo *models.Container) {
	scriptID := containerInfo.Scripts_id
	userName := user.Username
	userToken := ""
	if user.APIToken != "" {
		userToken = user.APIToken
	}
	//TODO 加密解密的问题
	containerInfo.Command = []string{fmt.Sprintf("/app/%s", scriptID)}
	containerInfo.Env = append(containerInfo.Env, []apiv1.EnvVar{
		{
			Name:  "SCRIPT_ENTRY_INFO",
			Value: scriptID + ":" + userName + ":" + userToken,
		},
		{
			Name:  "SCRIPT_URL",
			Value: common.ScriptUrl,
		},
	}...)

}

//状态。0-成功 1-失败 2-执行中 3-等待 子任务     flow 状态。0-成功 1-失败 2-执行中
// update build status with 'currentBuild' and start next build of same stage
type BuildStageOptions struct {
	BuildWithDependency bool
	FlowOwner           string
	ImageName           string
	UseCustomRegistry   bool
}

func HandleWaitTimeout(job *v1.Job, imageBuilder *models.ImageBuilder, stageBuildId string) (pod apiv1.Pod, timeout bool, err error) {
	method := "handleWaitTimeout"

	pod, err = imageBuilder.GetPod(job.ObjectMeta.Namespace, job.ObjectMeta.Name, stageBuildId)
	if err != nil {
		glog.Errorf("%s get %s pod failed:%v\n", method, pod.ObjectMeta.Name, err)
	}

	glog.Infof("%s - podName=[%s]<<===============>>", method, pod.ObjectMeta.Name)

	for i := 0; i < 5; i++ {
		time.Sleep(5 * time.Second)
		if pod.ObjectMeta.Name != "" {

			glog.Infof("%s - Checking if scm container is timeout\n", method)
			if len(pod.Status.InitContainerStatuses) > 0 &&
				IsContainerCreated(imageBuilder.ScmName, pod.Status.InitContainerStatuses) {
				glog.Infof("ContainerStatuses=========imageBuilder.BuilderName=:%s\n", imageBuilder.ScmName)
				timeout = false
				return
			}

			glog.Infof("%s - Checking if build container is timeout\n", method)
			if len(pod.Status.ContainerStatuses) > 0 &&
				IsContainerCreated(imageBuilder.BuilderName, pod.Status.ContainerStatuses) {
				glog.Infof("ContainerStatuses=========imageBuilder.BuilderName=:%s\n", imageBuilder.BuilderName)
				timeout = false
				return
			}

		}
	}

	//终止job
	glog.Infof("%s - will stop job=[%s]\n", method, job.ObjectMeta.Name)
	//1 代表手动停止 0表示程序停止
	//_, err = imageBuilder.StopJob(job.ObjectMeta.Namespace, job.ObjectMeta.Name, false, 0)
	//if err != nil {
	//	glog.Errorf("%s Stop the job %s failed: %v\n", method, job.ObjectMeta.Name, err)
	//}
	timeout = true
	return

}

func IsContainerCreated(ContainerName string, containerStatuses []apiv1.ContainerStatus) bool {

	for _, containerstatus := range containerStatuses {
		if ContainerName == containerstatus.Name {
			glog.Infof("The container %s status:%v\n", ContainerName, containerstatus.State)
			// 判断builder容器是否存在或是否重启过，从而判断是否容器创建成功
			if containerstatus.ContainerID != "" || containerstatus.RestartCount > 0 || containerstatus.State.Waiting != nil {
				glog.Infof("The container %s status:%v\n", ContainerName, containerstatus.State)
				return true
			}
		}

	}
	return false
}
