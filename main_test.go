package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestHelloCommand(t *testing.T) {
	// 临时替换 os.Args 用于模拟命令行输入
	//args := []string{"dolctl", "mysql", "list-backup", "--cluster-sign", "standard", "--namespace", "kedacom-project-namespace", "--name", "dol-mysql"}
	//args := []string{"dolctl", "mysql", "backup", "--cluster-sign", "standard", "--namespace", "kedacom-project-namespace", "--name", "dol-mysql", "--database", "all"}
	//args := []string{"dolctl", "mysql", "restore", "--cluster-sign", "standard", "--src-namespace", "kedacom-project-namespace", "--src-name", "dol-mysql", "--dest-namespace", "kedacom-project-namespace", "--dest-name", "dol-mysql", "--file", "/backup/20250117/clog/20250117213207_clog.sql.gz"}
	//args := []string{"dolctl", "tidb", "list-backup", "--cluster-sign", "standard", "--namespace", "kedacom-project-namespace", "--name", "dol-tidb"}
	//args := []string{"dolctl", "tidb", "backup", "--cluster-sign", "standard", "--namespace", "kedacom-project-namespace", "--name", "dol-tidb", "--database", "all"}
	//args := []string{"dolctl", "tidb", "restore", "--cluster-sign", "standard", "--src-namespace", "kedacom-project-namespace", "--src-name", "dol-tidb", "--dest-namespace", "kedacom-project-namespace", "--dest-name", "dol-tidb", "--file", "/backup/20250117/test"}
	args := []string{"devops-tool", "cluster", "get-pv", "--file", "D:\\go-script\\github\\devops-tools\\2.xlsx"}

	os.Args = args

	// 通过执行命令，捕获输出
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// 执行命令
	if err := Execute(rootCmd); err != nil {
		fmt.Print(err)
	}

}
