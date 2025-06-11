package cluster

import (
	"context"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"text/tabwriter"
	"time"
)

func GetStorageClassInfo(client *kubernetes.Clientset, filePath string) error {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	storageClassList, err := client.StorageV1().StorageClasses().List(ctx, metaV1.ListOptions{})
	if err != nil {
		return err
	}
	// 控制台输出表格
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	if filePath == "" {
		fmt.Fprintln(w, "NAME\tPROVISIONER\tRECLAIM POLICY")
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
		row.WriteSlice([]interface{}{"NAME", "PROVISIONER", "RECLAIM POLICY"}, -1)
	}

	for _, sc := range storageClassList.Items {
		name := sc.Name
		provisioner := sc.Provisioner
		reclaimPolicy := "Delete"
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		// 控制台打印
		if filePath == "" {
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, provisioner, reclaimPolicy)
		}

		// 如果指定了文件路径，则写入 Excel
		if filePath != "" {
			row := sheet.AddRow()
			row.WriteSlice([]interface{}{name, provisioner, reclaimPolicy}, -1)
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
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	pvList, err := client.CoreV1().PersistentVolumes().List(ctx, metaV1.ListOptions{})
	if err != nil {
		return err
	}

	// 控制台输出表格
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	if filePath == "" {
		fmt.Fprintln(w, "NAME\tCAPACITY\tACCESS MODES\tRECLAIM POLICY\tSTATUS\tCLAIM\tSTORAGECLASS\tTYPE\tLOCATION\tAGE")
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
			"CLAIM", "STORAGECLASS", "TYPE", "LOCATION", "AGE",
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
		if pv.Spec.ClaimRef != nil {
			claim = fmt.Sprintf("%s/%s/%s", pv.Spec.ClaimRef.Kind, pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		// 初始化 TYPE 和 LOCATION
		pvType := "unknown"
		location := ""

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
								location = fmt.Sprintf("%s:%s", req.Values[0], path)
							}
						}
					}
				}
			}
		} else if pv.Spec.PersistentVolumeSource.CephFS != nil {
			pvType = "ceph"
			location = fmt.Sprintf("%s:%s", pv.Spec.PersistentVolumeSource.CephFS.Monitors[0], pv.Spec.PersistentVolumeSource.CephFS.Path) // 取第一个 monitor 示例
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
			fmt.Fprintf(w, "%s\t%s\t%v\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				name, capacity, accessModes, reclaimPolicy, status, claim, storageClass, pvType, location, age)
		}

		// 如果指定了文件路径，则写入 Excel
		if filePath != "" {
			row := sheet.AddRow()
			row.WriteSlice([]interface{}{
				name, capacity, fmt.Sprintf("%v", accessModes), reclaimPolicy, status,
				claim, storageClass, pvType, location, age,
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
