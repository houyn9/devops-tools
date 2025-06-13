package cluster

import (
	"context"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	appsv1 "k8s.io/api/apps/v1"
	bv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func GetStorageClassInfo(client *kubernetes.Clientset, filePath string) error {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	storageClassList, err := client.StorageV1().StorageClasses().List(ctx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	// 控制台输出表格
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	if filePath == "" {
		fmt.Fprintln(w, "NAME\tPROVISIONER\tRECLAIM POLICY\tNAMESPACE BOUND")
	}

	// 创建 Excel 文件（如果 filePath 非空）
	var file *xlsx.File
	var sheet *xlsx.Sheet
	if filePath != "" {
		file = xlsx.NewFile()
		sheet, err = file.AddSheet("StorageClasses")
		if err != nil {
			return err
		}
		// 添加表头
		row := sheet.AddRow()
		row.WriteSlice([]interface{}{"NAME", "PROVISIONER", "RECLAIM POLICY", "NAMESPACE BOUND"}, -1)
	}

	for _, sc := range storageClassList.Items {
		name := sc.Name
		provisioner := sc.Provisioner
		reclaimPolicy := "Delete"
		namespacesBound := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}
		for _, item := range namespaces.Items {
			if item.Annotations != nil {
				for _, s := range strings.Split(item.Annotations["dophin/storage"], ",") {
					if s == sc.Name {
						namespacesBound += item.Name + ","
					}
				}
			}
		}

		// 控制台打印
		if filePath == "" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, provisioner, reclaimPolicy, namespacesBound)
		}

		// 如果指定了文件路径，则写入 Excel
		if filePath != "" {
			row := sheet.AddRow()
			row.WriteSlice([]interface{}{name, provisioner, reclaimPolicy, namespacesBound}, -1)
		}
	}
	w.Flush()

	// 保存 Excel 文件
	if filePath != "" {
		if err := file.Save(filePath); err != nil {
			return err
		}
		fmt.Printf("StorageClass 数据已写入文件: %s\n", filePath)
	}

	return nil
}
func GetPersistentVolumeInfo(client *kubernetes.Clientset, filePath string) error {
	pvCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	pvList, err := client.CoreV1().PersistentVolumes().List(pvCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	nodeCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	nodeList, err := client.CoreV1().Nodes().List(nodeCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	nodeMap := make(map[string]bool)
	for _, item := range nodeList.Items {
		nodeMap[item.Name] = true
	}
	podCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	podList, err := client.CoreV1().Pods("").List(podCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	deploymentCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	deploymentList, err := client.AppsV1().Deployments("").List(deploymentCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	daemonSetCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	daemonSetList, err := client.AppsV1().DaemonSets("").List(daemonSetCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	statefulSetCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	statefulSetList, err := client.AppsV1().StatefulSets("").List(statefulSetCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	cronJobCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	cronJobList, err := client.BatchV1().CronJobs("").List(cronJobCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	jobCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	jobList, err := client.BatchV1().Jobs("").List(jobCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	pvcCtx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	pvcList, err := client.CoreV1().PersistentVolumeClaims("").List(pvcCtx, metaV1.ListOptions{})
	if err != nil {
		return err
	}

	// 控制台输出表格
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	if filePath == "" {
		fmt.Fprintln(w, "NAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGECLASS\tTYPE\tLOCATION\tAGE\tNODE_ISEXIST\tBONDPVCISEXIST\tPVCINUSE")
	}

	// 创建 Excel 文件（如果 filePath 非空）
	var file *xlsx.File
	var sheet *xlsx.Sheet
	if filePath != "" {
		file = xlsx.NewFile()
		sheet, err = file.AddSheet("PersistentVolumes")
		if err != nil {
			return err
		}
		// 添加表头
		row := sheet.AddRow()
		row.WriteSlice([]interface{}{
			"NAME", "CAPACITY", "ACCESS MODES", "RECLAIM POLICY", "STATUS",
			"CLAIM", "STORAGECLASS", "TYPE", "LOCATION", "AGE", "NODEISEXIST", "BONDPVCISEXIST", "PVCINUSE",
		}, -1)
	}

	for _, pv := range pvList.Items {
		name := pv.Name
		capacity := pv.Spec.Capacity.Storage().String()
		accessModes := pv.Spec.AccessModes
		reclaimPolicy := string(pv.Spec.PersistentVolumeReclaimPolicy)
		status := pv.Status.Phase
		storageClass := pv.Spec.StorageClassName
		age := time.Since(pv.CreationTimestamp.Time).Round(time.Second)
		// 提取 CLAIM 字段
		claim := ""
		pvcInUse := false
		bondPvcIsExist := false
		if pv.Spec.ClaimRef != nil {
			claim = fmt.Sprintf("%s/%s/%s", pv.Spec.ClaimRef.Kind, pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
			for _, item := range pvcList.Items {
				if item.Namespace == pv.Spec.ClaimRef.Namespace && item.Name == pv.Spec.ClaimRef.Name && item.UID == pv.Spec.ClaimRef.UID {
					bondPvcIsExist = true
					pvcInUse, err = isPVCUsed(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, podList, deploymentList, statefulSetList, daemonSetList, cronJobList, jobList)
					if err != nil {
						return err
					}
					break
				}
			}
		}

		// 初始化 TYPE 和 LOCATION
		pvType := "unknown"
		location := ""
		nodeIsExist := ""

		// 判断 PV 类型，并提取对应路径或服务器信息
		if pv.Spec.PersistentVolumeSource.Local != nil {
			pvType = "local"
			for key, value := range pv.Labels {
				if key == "dolphin.storage/sc-type" && value == "sig-local" {
					pvType = "shard_local"
				}
			}
			path := pv.Spec.PersistentVolumeSource.Local.Path
			// 检查 NodeAffinity 并拼接节点信息
			if affinity := pv.Spec.NodeAffinity; affinity != nil {
				if term := affinity.Required.NodeSelectorTerms; len(term) > 0 {
					for _, t := range term {
						for _, req := range t.MatchExpressions {
							if req.Key == "kubernetes.io/hostname" && len(req.Values) > 0 {
								location = fmt.Sprintf("%s:%s", strings.Join(req.Values, ","), path)
								nodeIsExist = "no"
								for _, value := range req.Values {
									if nodeMap[value] {
										nodeIsExist = "yes"
									}
								}
							}
						}
					}
				}
			}
		} else if pv.Spec.PersistentVolumeSource.CephFS != nil {
			pvType = "ceph"
			location = fmt.Sprintf("%s:%s", strings.Join(pv.Spec.PersistentVolumeSource.CephFS.Monitors, ","), pv.Spec.PersistentVolumeSource.CephFS.Path) // 取第一个 monitor 示例
		} else if pv.Spec.PersistentVolumeSource.NFS != nil {
			pvType = "nfs"
			location = fmt.Sprintf("%s:%s", pv.Spec.PersistentVolumeSource.NFS.Server, pv.Spec.PersistentVolumeSource.NFS.Path)
		} else if pv.Spec.PersistentVolumeSource.HostPath != nil {
			pvType = "hostpath"
			location = pv.Spec.PersistentVolumeSource.HostPath.Path
		} else {
			// 其他类型如云盘等可根据需要扩展
		}

		// 控制台打印
		if filePath == "" {
			fmt.Fprintf(w, "%s\t%s\t%v\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%v\t%v\n",
				name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, pvType, location, age, nodeIsExist, bondPvcIsExist, pvcInUse)
		}

		// 如果指定了文件路径，则写入 Excel
		if filePath != "" {
			row := sheet.AddRow()
			row.WriteSlice([]interface{}{
				name, capacity, fmt.Sprintf("%v", accessModes), reclaimPolicy, status,
				claim, storageClass, pvType, location, age, nodeIsExist, bondPvcIsExist, pvcInUse,
			}, -1)
		}
	}
	w.Flush()

	// 保存 Excel 文件
	if filePath != "" {
		if err := file.Save(filePath); err != nil {
			return err
		}
		fmt.Printf("PersistentVolume 数据已写入文件: %s\n", filePath)
	}

	return nil
}

func isPVCUsed(namespace, pvcName string, podList *corev1.PodList, dpList *appsv1.DeploymentList, stsList *appsv1.StatefulSetList, dsList *appsv1.DaemonSetList, cjList *bv1.CronJobList, jobList *bv1.JobList) (bool, error) {
	used := false

	// 检查 Pod
	for _, pod := range podList.Items {
		if pod.Namespace != namespace {
			continue
		}
		if isPVCInVolumes(pod.Spec.Volumes, pvcName) {
			used = true
			goto RETURNRESULT
		}
	}

	// 检查 Deployment
	for _, deploy := range dpList.Items {
		if deploy.Namespace != namespace {
			continue
		}
		if isPVCInVolumes(deploy.Spec.Template.Spec.Volumes, pvcName) {
			used = true
			goto RETURNRESULT
		}
	}

	// 检查 StatefulSet
	for _, sts := range stsList.Items {
		if sts.Namespace != namespace && !strings.Contains(pvcName, "-") {
			continue
		}
		if sts.Spec.VolumeClaimTemplates != nil {
			for _, pvc := range sts.Spec.VolumeClaimTemplates {
				if pvc.Name+"-"+sts.Name == pvcName[:strings.LastIndex(pvcName, "-")] {
					used = true
					goto RETURNRESULT
				}
			}
		}
	}

	// 检查 DaemonSet
	for _, ds := range dsList.Items {
		if ds.Namespace != namespace {
			continue
		}
		if isPVCInVolumes(ds.Spec.Template.Spec.Volumes, pvcName) {
			used = true
			goto RETURNRESULT
		}
	}

	// 检查 ReplicaSet
	//for _, rs := range rsList.Items {
	//	if isPVCInVolumes(rs.Spec.Template.Spec.Volumes, pvcName) {
	//		refDetails = append(refDetails, "ReplicaSet/"+rs.Name)
	//		used = true
	//	}
	//}

	// 检查 Job
	for _, job := range jobList.Items {
		if job.Status.CompletionTime != nil {
			continue
		}
		if isPVCInVolumes(job.Spec.Template.Spec.Volumes, pvcName) {
			used = true
			goto RETURNRESULT
		}
	}

	// 检查 CronJob
	for _, cj := range cjList.Items {
		if cj.Namespace != namespace {
			continue
		}
		if isPVCInVolumes(cj.Spec.JobTemplate.Spec.Template.Spec.Volumes, pvcName) {
			used = true
			goto RETURNRESULT
		}
	}
RETURNRESULT:
	return used, nil
}

func isPVCInVolumes(volumes []corev1.Volume, pvcName string) bool {
	for _, v := range volumes {
		if v.PersistentVolumeClaim != nil && v.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}

//
//var metav1ListOpts = metaV1.ListOptions{}
