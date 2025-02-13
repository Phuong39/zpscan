package webscan

import (
	"github.com/projectdiscovery/gologger"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
)

var (
	/*
		<head>
			<meta http-equiv="refresh" content="1;URL='/admin'"/>
		</head>

		<!--
		<meta http-equiv="refresh" content="0.1;url=https://www.xxx.cn/">
		-->

		<head>
		<meta http-equiv=refresh content=0;url=index.jsp>
		</head>
	*/
	reg1 = regexp.MustCompile(`(?i)<meta.*?http-equiv=.*?refresh.*?url=(.*?)/?>`)
	/*
		<script type="text/javascript">
			location.href = "./ui/";
		</script>
	*/
	reg2 = regexp.MustCompile(`(?i)location\.href.*?=.*?["'](.*?)["']`)
	/*
		<script language="javascript">
			window.location.replace("/mymeetings/");
		</script>
	*/
	reg3 = regexp.MustCompile(`(?i)window\.location\.replace\(['"](.*?)['"]\)`)
)

func Jsjump(resp *req.Response) (jumpurl string) {
	res := regexJsjump(resp)
	if res != "" {
		gologger.Debug().Msgf("regexJsjump(), %v", res)
		res = strings.TrimSpace(res)
		res = strings.ReplaceAll(res, "\"", "")
		res = strings.ReplaceAll(res, "'", "")
		res = strings.ReplaceAll(res, "./", "/")
		if strings.HasPrefix(res, "http") {
			jumpurl = res
		} else if strings.HasPrefix(res, "/") {
			// 前缀存在 / 时拼接绝对目录
			baseUrl := resp.Request.URL.Scheme + "://" + resp.Request.URL.Host
			jumpurl = baseUrl + res
		} else {
			// 前缀不存在 / 时拼接相对目录
			baseUrl := resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + "/" + filepath.Dir(resp.Request.URL.Path) + "/"
			jumpurl = baseUrl + res
		}
	}
	return
}

func regexJsjump(resp *req.Response) string {
	matches := reg1.FindAllStringSubmatch(resp.String(), -1)
	if len(matches) > 0 {
		// 去除注释的情况
		if !strings.Contains(resp.String(), "<!--\r\n"+matches[0][0]) && !strings.Contains(matches[0][1], "nojavascript.html") && !strings.Contains(resp.String(), "<!--[if lt IE 7]>\n"+matches[0][0]) {
			return matches[0][1]
		}
	}
	body := resp.String()
	if len(body) > 700 {
		body = body[:700]
	}
	matches = reg2.FindAllStringSubmatch(body, -1)
	if len(matches) > 0 {
		return matches[0][1]
	}
	matches = reg3.FindAllStringSubmatch(body, -1)
	if len(matches) > 0 {
		return matches[0][1]
	}
	return ""
}
