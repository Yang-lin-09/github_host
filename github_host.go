package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"bufio"
	"time"
	"github.com/PuerkitoBio/goquery"
)

type GithubHost struct {
	Domain string
	IP string
	Err error
}

var domains = []string {
	"github.githubassets.com",
	"central.github.com",
	"desktop.githubusercontent.com",
	"assets-cdn.github.com",
	"camo.githubusercontent.com",
	"github.map.fastly.net",
	"github.global.ssl.fastly.net",
	"gist.github.com",
	"github.io",
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
	"user-images.githubusercontent.com",
	"favicons.githubusercontent.com",
	"avatars5.githubusercontent.com",
	"avatars4.githubusercontent.com",
	"avatars3.githubusercontent.com",
	"avatars2.githubusercontent.com",
	"avatars1.githubusercontent.com",
	"avatars0.githubusercontent.com",
	"avatars.githubusercontent.com",
	"codeload.github.com",
	"github-cloud.s3.amazonaws.com",
	"github-com.s3.amazonaws.com",
	"github-production-release-asset-2e65be.s3.amazonaws.com",
	"github-production-user-asset-6210df.s3.amazonaws.com",
	"github-production-repository-file-5c1aeb.s3.amazonaws.com",
	"githubstatus.com",
	"github.community",
	"media.githubusercontent.com",
}

var start_tag = "# github_host start\n"
var end_tag = "# github_host end\n"

func main() {

	file_path := "./hosts"
	host_map := get_host()

	time_unix := time.Now().Unix() 
	str := time.Unix(time_unix, 0).Format("2006-01-02 15:04:05")
	str = "# update on " + str + "\n"
	str += start_tag
	for _, domain := range domains {
		str += host_map[domain] + " " + domain + "\n"
	}
	str += end_tag

	out, err := os.OpenFile(file_path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open file fail:", err)
		return
	}
	defer out.Close()
	writer := bufio.NewWriter(out)
	writer.WriteString(str)
	writer.Flush()
	fmt.Println("write file success, position: ", file_path)
}

func get_host() map[string]string {
	ch := make(chan *GithubHost)
	for _, domain := range domains {
		go http_post_form(domain, ch)
	}

	host_map := make(map[string]string)
	for range domains {
		ch_rec := <-ch;
		if ch_rec.Err != nil {
			fmt.Println(ch_rec.Err.Error() + " " + ch_rec.Domain)
		}

		host_map[ch_rec.Domain] = ch_rec.IP
		fmt.Println(ch_rec.IP + " " + ch_rec.Domain)
	}

	return host_map
}

func http_post_form(domain string, ch chan<- *GithubHost) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.ipaddress.com/ip-lookup", strings.NewReader("host=" + domain))
	if err != nil {
		ch <- &GithubHost{Domain: domain, Err: err}
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		ch <- &GithubHost{Domain: domain, Err: err}
		return
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		ch <- &GithubHost{Domain: domain, Err: err}
		return
	}

	title := doc.Find("title").Text()
	ip := find_ipv4(title)

	if ip == "" {
		text := doc.Find(".resp main").Find("form:first-of-type").Find("div:first-child").Find("a").Text()
		ip = find_ipv4(text)

		if ip == "" {
			err = fmt.Errorf("can not find ipv4 address")
			ch <- &GithubHost{Domain: domain, Err: err}
			return
		}
	}

	ch <- &GithubHost{Domain: domain, IP: ip, Err: nil}
}

func find_ipv4(input string) string {
	part_ip := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	grammer := part_ip + "\\." + part_ip + "\\." + part_ip + "\\." + part_ip
	match_me := regexp.MustCompile(grammer)
	return match_me.FindString(input)
}