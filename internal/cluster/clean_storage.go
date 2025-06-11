package cluster

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
	"time"
)

const (
	BackupScDir = "/data/storage-clean/sc"
	BackupPvDir = "/data/storage-clean/pv"
	LogFile     = "/data/storage-clean/clean.log"
)

var (
	scheme      = runtime.NewScheme()
	currentTime = time.Now().Format("2006-01-02-15:04:05")
)

func init() {
	_ = corev1.AddToScheme(scheme)
	_ = storagev1.AddToScheme(scheme)
}

// CleanStorageResources 清理集群中的 StorageClass 和 PV 资源
func CleanStorageResources(client *kubernetes.Clientset) error {
	if err := os.MkdirAll("/data/storage-clean", 0755); err != nil {
		return fmt.Errorf("创建备份目录失败/data/storage-clean: %v", err)
	}
	logToFile("开始执行存储资源清理任务...\n")

	if err := deleteUnusedStorageClasses(client); err != nil {
		logToFile("清理 StorageClass 出错: %v\n", err)
		return err
	}

	if err := cleanupPersistentVolumes(client); err != nil {
		logToFile("清理 PV 出错: %v\n", err)
		return err
	}

	logToFile("存储资源清理完成。\n")
	return nil
}
func deleteUnusedStorageClasses(client *kubernetes.Clientset) error {
	scList, err := client.StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	usedSC := make(map[string]bool)
	for _, pv := range pvList.Items {
		if pv.Spec.StorageClassName != "" {
			usedSC[pv.Spec.StorageClassName] = true
		}
	}

	for _, sc := range scList.Items {
		if !usedSC[sc.Name] {
			logToFile("准备删除未使用的 StorageClass: %s\n", sc.Name)
			err := backupResource(&sc, BackupScDir+currentTime)
			if err != nil {
				logToFile("备份 StorageClass %s 失败: %v\n", sc.Name, err)
			}
			if err := client.StorageV1().StorageClasses().Delete(context.Background(), sc.Name, metav1.DeleteOptions{}); err != nil {
				logToFile("删除 StorageClass %s 失败: %v\n", sc.Name, err)
				continue
			}
			logToFile("成功删除并备份 StorageClass: %s\n", sc.Name)
		}
	}
	return nil
}
func cleanupPersistentVolumes(client *kubernetes.Clientset) error {
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pv := range pvList.Items {
		switch pv.Status.Phase {
		case "Available":
			logToFile("PV %s 状态为 Available，准备删除并备份\n", pv.Name)
			if err := backupResource(&pv, BackupPvDir+currentTime); err != nil {
				logToFile("备份 PV %s 失败: %v\n", pv.Name, err)
			}
			if err := client.CoreV1().PersistentVolumes().Delete(context.Background(), pv.Name, metav1.DeleteOptions{}); err != nil {
				logToFile("删除 PV %s 失败: %v\n", pv.Name, err)
			} else {
				logToFile("成功删除并备份 PV: %s\n", pv.Name)
			}
		case "Released":
			var pvcNamespace, pvcName string
			if ref := pv.Spec.ClaimRef; ref != nil {
				pvcNamespace = ref.Namespace
				pvcName = ref.Name
			} else {
				logToFile("PV %s 状态为 Released，但无 ClaimRef，直接删除\n", pv.Name)
				if err := backupResource(&pv, BackupPvDir+currentTime); err != nil {
					logToFile("备份 PV %s 失败: %v\n", pv.Name, err)
				}
				if err := client.CoreV1().PersistentVolumes().Delete(context.Background(), pv.Name, metav1.DeleteOptions{}); err != nil {
					logToFile("删除 PV %s 失败: %v\n", pv.Name, err)
				} else {
					logToFile("成功删除并备份 PV: %s\n", pv.Name)
				}
				continue
			}

			pvcInfo, err := client.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(context.Background(), pvcName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					logToFile("PVC %s/%s 不存在，准备删除 PV %s\n", pvcNamespace, pvcName, pv.Name)
					if err := backupResource(&pv, BackupPvDir+currentTime); err != nil {
						logToFile("备份 PV %s 失败: %v\n", pv.Name, err)
					}
					if err := client.CoreV1().PersistentVolumes().Delete(context.Background(), pv.Name, metav1.DeleteOptions{}); err != nil {
						logToFile("删除 PV %s 失败: %v\n", pv.Name, err)
					} else {
						logToFile("成功删除并备份 PV: %s\n", pv.Name)
					}
				} else {
					logToFile("获取 PVC %s/%s 异常: %v\n", pvcNamespace, pvcName, err)
				}
			} else {
				if pv.Spec.ClaimRef.UID != "" && pv.Spec.ClaimRef.UID != pvcInfo.UID {
					logToFile("PVC %s/%s 存在，但 UID 不匹配，准备删除 PV %s\n", pvcNamespace, pvcName, pv.Name)
					if err := backupResource(&pv, BackupPvDir+currentTime); err != nil {
						logToFile("备份 PV %s 失败: %v\n", pv.Name, err)
					}
					if err := client.CoreV1().PersistentVolumes().Delete(context.Background(), pv.Name, metav1.DeleteOptions{}); err != nil {
						logToFile("删除 PV %s 失败: %v\n", pv.Name, err)
					} else {
						logToFile("成功删除并备份 PV: %s\n", pv.Name)
					}
				} else {
					logToFile("PV %s 正在被 PVC 使用，跳过删除\n", pv.Name)
				}
			}
		default:
			logToFile("PV %s 状态为 %s，跳过删除\n", pv.Name, pv.Status.Phase)
		}
	}
	return nil
}
func backupResource(obj runtime.Object, backupDir string) error {
	// 创建序列化器
	yamlSerializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)

	// 从对象中提取元数据
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("获取对象元数据失败: %v", err)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Empty() {
		// 尝试从注册的 scheme 中识别 GVK
		gvks, _, err := scheme.ObjectKinds(obj)
		if err != nil || len(gvks) == 0 {
			return fmt.Errorf("无法从 scheme 获取 GVK: %v", err)
		}
		gvk = gvks[0]
		obj.GetObjectKind().SetGroupVersionKind(gvk)
	}

	// 构建备份路径
	fileName := fmt.Sprintf("%s-%s.yaml", gvk.Kind, accessor.GetName())
	filePath := filepath.Join(backupDir, fileName)

	// 确保目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 打开文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}
	defer file.Close()

	// 执行序列化
	if err := yamlSerializer.Encode(obj, file); err != nil {
		return fmt.Errorf("序列化资源对象失败: %v", err)
	}

	return nil
}
func logToFile(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开日志文件: %v\n", err)
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format("[2006-01-02 15:04:05] ") + msg)
}
