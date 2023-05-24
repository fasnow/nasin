package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"nasin/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	banner = "   ____   ____ _ _____ (_)____ \n" +
		"  / __ \\ / __ `// ___// // __ \\\n" +
		" / / / // /_/ /(__  )/ // / / /\n" +
		"/_/ /_/ \\__,_//____//_//_/ /_/ \n" +
		"    @github.com/fasnow"
	usage = `  -t string                                      
        target http://host:port/ without path    
  -u string                                      
        nacos username
  -p string                                      
        nacos password 
  -o string                                      
        output file, a+ mod (default "nasin.txt") 
  -print                                         
        print config detail when processing      
  -proxy string                                  
        socks5://host:port or http://host:port`
)

type FlagConfig struct {
	proxy       string
	target      string
	pass        string
	user        string
	output      string
	printDetail bool
}

func (c *FlagConfig) ParseFlags() {
	fmt.Println(banner)
	flag.StringVar(&c.target, "t", "", "target http://host:port/ without path")
	flag.StringVar(&c.user, "u", "", "nacos username")
	flag.StringVar(&c.pass, "p", "", "nacos password")
	flag.StringVar(&c.output, "o", "nasin.txt", "output file, a+ mod")
	flag.BoolVar(&c.printDetail, "print", false, "print config detail when processing")
	flag.StringVar(&c.proxy, "proxy", "", "socks5://host:port or http://host:port")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: \n  %s [options]\n", filepath.Base(os.Args[0]))
		fmt.Println("Options:")
		//flag.PrintDefaults()
		fmt.Print(usage)
	}
	flag.Parse()
	var set []string
	if c.target == "" {
		set = append(set, "target")
	}
	if c.user == "" {
		set = append(set, "user")
	}
	if c.pass == "" {
		set = append(set, "pass")
	}
	if set != nil || len(set) != 0 {
		fmt.Println(fmt.Sprintf("Error: required flag(s) %s not set", strings.Join(set, ", ")))
		flag.Usage()
		os.Exit(0)
	}
	if c.proxy != "" {
		http.SetGlobalProxy(c.proxy)
	}

	execNasin(*c)
}

func execNasin(c FlagConfig) {
	client, err := http.NewNasinClient(c.target, c.user, c.pass)
	if err != nil {
		fmt.Println(err)
		return
	}
	projectList, err := client.GetProject()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(projectList) == 0 {
		fmt.Println("project is empty")
		return
	}
	fmt.Println("available projects")
	for _, v := range projectList {
		content := fmt.Sprintf("\tname: %s    namespace: %s    configCount: %d", v.NamespaceShowName, v.Namespace, v.ConfigCount)
		fmt.Println(content)
		save(c.output, content)
	}
	for _, project := range projectList {
		fmt.Println(fmt.Sprintf("is getting %s config", project.NamespaceShowName))
		if project.ConfigCount == 0 {
			//不处理会报500错误
			fmt.Println(fmt.Sprintf("%s's config is empty", project.NamespaceShowName))
			content := fmt.Sprintf("======%s====%s\n%s\n", project.NamespaceShowName, "config is empty", "")
			save(c.output, content)
			continue
		}
		configList, err := client.GetProjectConfig(project)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Println(fmt.Sprintf("%s's available config", project.NamespaceShowName))
		for _, v := range configList {
			fmt.Println("\t" + v.Name)
		}
		for _, config := range configList {
			if c.printDetail {
				fmt.Println(project.NamespaceShowName + " detail")
				fmt.Println(config.Content)
			}
			content := fmt.Sprintf("======%s====%s\n%s\n", project.NamespaceShowName, config.Name, config.Content)
			save(c.output, content)
		}
	}
	fmt.Println("Finished")
}

func save(fileName, content string) {
	// 打开文件，如果文件不存在则创建，以追加模式写入
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing writer:", err)
	}
}
