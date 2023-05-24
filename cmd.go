package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	proxy       string
	target      string
	pass        string
	user        string
	output      string
	printDetail bool
	banner      = "   ____  ____ ______(_)___ \n" +
		"  / __ \\/ __ `/ ___/ / __ \\\n" +
		" / / / / /_/ (__  ) / / / /\n" +
		"/_/ /_/\\__,_/____/_/_/ /_/ \n " +
		"  @github.com/fasnow   "
)

var rootCmd = &cobra.Command{
	Use:   "nasin",
	Short: "Description: download nacos config list",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var set []string
		if target == "" {
			set = append(set, "target")
		}
		if user == "" {
			set = append(set, "user")
		}
		if pass == "" {
			set = append(set, "pass")
		}
		if set != nil || len(set) != 0 {
			fmt.Println(fmt.Sprintf("Error: required flag(s) %s not set", strings.Join(set, ", ")))
			cmd.Usage()
			os.Exit(0)
		}
		if proxy != "" {
			SetGlobalProxy(proxy)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		execNasin()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&target, "target", "t", "", "http://host:port/ with no path")
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "username")
	rootCmd.Flags().StringVarP(&pass, "pass", "p", "", "password")
	rootCmd.Flags().StringVarP(&output, "output", "o", "nasin.txt", "output file, a+ mod")
	rootCmd.Flags().BoolVar(&printDetail, "print", false, "print config detail when processing")
	rootCmd.Flags().StringVarP(&pass, "checkVersion", "v", "", "check new version")
	rootCmd.Flags().StringVar(&proxy, "proxy", "", "socks5://host:port or http://host:port")
	fmt.Println(banner)
}

func execNasin() {
	client, err := NewNasinClient(target, user, pass)
	if err != nil {
		log.Println(err)
		return
	}
	projectList, err := client.GetProject()
	if err != nil {
		log.Println(err)
		return
	}
	if len(projectList) == 0 {
		log.Println("project is empty")
		return
	}
	log.Println("available projects")
	for _, v := range projectList {
		content := fmt.Sprintf("\tname: %s    namespace: %s    configCount: %d", v.NamespaceShowName, v.Namespace, v.ConfigCount)
		fmt.Println(content)
		save(content)
	}
	for _, project := range projectList {
		log.Println(fmt.Sprintf("is getting %s config", project.NamespaceShowName))
		if project.ConfigCount == 0 {
			//不处理会报500错误
			log.Println(fmt.Sprintf("%s's config is empty", project.NamespaceShowName))
			content := fmt.Sprintf("======%s====%s\n%s\n", project.NamespaceShowName, "config is empty", "")
			save(content)
			continue
		}
		configList, err := client.GetProjectConfig(project)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(fmt.Sprintf("%s's available config", project.NamespaceShowName))
		for _, v := range configList {
			fmt.Println("\t" + v.name)
		}
		for _, config := range configList {
			if printDetail {
				log.Println(project.NamespaceShowName + " detail")
				fmt.Println(config.content)
			}
			content := fmt.Sprintf("======%s====%s\n%s\n", project.NamespaceShowName, config.name, config.content)
			save(content)
		}
	}
	log.Println("Finished")
}

func save(content string) {
	// 打开文件，如果文件不存在则创建，以追加模式写入
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		log.Println("Error writing to file:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Println("Error flushing writer:", err)
		return
	}
}
