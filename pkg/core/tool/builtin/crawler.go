package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/DaWesen/Pandora/pkg/core"
	"github.com/PuerkitoBio/goquery"
)

// 实现网页爬虫工具
type CrawlerTool struct {
	client *http.Client
}

// 创建新的爬虫工具实例
func NewCrawlerTool() *CrawlerTool {
	return &CrawlerTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 返回工具名称
func (t *CrawlerTool) Name() string {
	return "crawler"
}

// 返回工具描述
func (t *CrawlerTool) Description() string {
	return "提供网页爬虫功能，爬取指定 URL 的网页内容并提取关键信息。"
}

// 返回工具参数模式
func (t *CrawlerTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"url": {
				Type:        "string",
				Description: "要爬取的网页 URL",
			},
			"extract": {
				Type:        "string",
				Description: "要提取的内容类型，如 'title'、'content'、'all'",
				Enum:        []string{"title", "content", "all"},
			},
			"output": {
				Type:        "string",
				Description: "输出格式，如 'txt'，默认为不存储",
				Enum:        []string{"txt"},
			},
			"filename": {
				Type:        "string",
				Description: "输出文件名，默认为自动生成",
			},
			"timeout": {
				Type:        "integer",
				Description: "爬取超时时间（秒），默认为 30",
			},
		},
		Required: []string{"url"},
	}
}

// 执行爬虫操作
func (t *CrawlerTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		URL      string `json:"url"`
		Extract  string `json:"extract,omitempty"`
		Output   string `json:"output,omitempty"`
		Filename string `json:"filename,omitempty"`
		Timeout  int    `json:"timeout,omitempty"`
	}

	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}

	if params.Extract == "" {
		params.Extract = "all"
	}
	if params.Timeout <= 0 {
		params.Timeout = 30
	}

	// 爬取网页
	result, err := t.crawl(params.URL, params.Extract)
	if err != nil {
		return core.ToolOutput{}, err
	}

	// 存储结果到文件
	if params.Output != "" {
		filename := params.Filename
		if filename == "" {
			// 自动生成文件名
			filename = fmt.Sprintf("crawl_result.%s", params.Output)
		} else {
			// 确保文件扩展名正确
			if !strings.HasSuffix(filename, "."+params.Output) {
				filename += "." + params.Output
			}
		}

		if err := t.saveResult(result, filename, params.Output); err != nil {
			return core.ToolOutput{}, err
		}

		return core.ToolOutput{
			Content: fmt.Sprintf("成功爬取 %s，提取了 %s 内容，并保存到 %s", params.URL, params.Extract, filename),
			Data:    result,
		}, nil
	}

	return core.ToolOutput{
		Content: fmt.Sprintf("成功爬取 %s，提取了 %s 内容", params.URL, params.Extract),
		Data:    result,
	}, nil
}

// 获取页面内容
func (t *CrawlerTool) FetchPage(pageURL string) (string, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP错误: %s", resp.Status)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	return string(body), nil
}

// 使用goquery精确提取内容
func (t *CrawlerTool) ExtractCleanContent(htmlContent string) (string, error) {
	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("解析HTML失败: %w", err)
	}

	var result strings.Builder

	// 提取标题
	title := doc.Find("title").Text()
	title = strings.Replace(title, " - 萌娘百科", "", 1)
	if title != "" {
		result.WriteString("页面标题: " + title + "\n")
		result.WriteString(strings.Repeat("=", 50) + "\n\n")
	}

	// 提取主要内容区域
	mainContent := doc.Find("#mw-content-text")
	if mainContent.Length() == 0 {
		mainContent = doc.Find(".mw-parser-output")
	}

	if mainContent.Length() > 0 {
		// 移除不需要的元素
		mainContent.Find("script, style, noscript, .navbox, .reference, .mw-editsection, .infobox").Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})

		// 提取段落文本
		mainContent.Find("p, h1, h2, h3, h4, h5, h6, li").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" && len(text) > 10 && !t.isUselessText(text) {
				// 清理文本
				text = t.cleanText(text)
				if text != "" {
					// 根据标签类型添加格式
					tagName := goquery.NodeName(s)
					switch tagName {
					case "h1":
						result.WriteString("\n" + text + "\n")
						result.WriteString(strings.Repeat("=", len(text)) + "\n\n")
					case "h2":
						result.WriteString("\n" + text + "\n")
						result.WriteString(strings.Repeat("-", len(text)) + "\n\n")
					case "h3", "h4", "h5", "h6":
						result.WriteString("\n" + text + "\n\n")
					default:
						result.WriteString(text + "\n\n")
					}
				}
			}
		})

		// 如果段落提取的内容太少，尝试提取整个文本
		if result.Len() < 1000 {
			fullText := strings.TrimSpace(mainContent.Text())
			fullText = t.cleanText(fullText)
			if len(fullText) > result.Len() {
				result.Reset()
				if title != "" {
					result.WriteString("页面标题: " + title + "\n")
					result.WriteString(strings.Repeat("=", 50) + "\n\n")
				}
				result.WriteString(fullText)
			}
		}
	} else {
		// 备用方法：提取body文本
		fullText := strings.TrimSpace(doc.Find("body").Text())
		fullText = t.cleanText(fullText)
		result.WriteString(fullText)
	}

	content := strings.TrimSpace(result.String())
	if content == "" {
		return "", fmt.Errorf("未能提取到有效内容")
	}

	return content, nil
}

// 清理文本
func (t *CrawlerTool) cleanText(text string) string {
	// 移除HTML实体
	text = t.decodeHTMLEntities(text)

	// 移除多余的空白字符
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// 移除常见的无用文本模式
	uselessPatterns := []string{
		`\[编辑\]`, `\[\d+\]`, `跳转到.*`, `分类.*`, `页面.*`, `讨论.*`,
		`查看.*`, `工具.*`, `登录.*`, `创建账户.*`, `萌娘百科.*`,
		`贡献.*`, `查看历史.*`, `搜索.*`, `导航.*`, `首页.*`,
		`最近更改.*`, `随机页面.*`, `帮助.*`, `相关更改.*`,
		`特殊页面.*`, `打印版本.*`, `永久链接.*`, `页面信息.*`,
		`引用本页.*`, `本文介绍的是.*`, `关于.*`, `参见.*`,
		`use strict.*`, `window\.RLQ\.push.*`, `document\.querySelector.*`,
		`function.*`, `const.*`, `let.*`, `var.*`, `return.*`,
		`if.*`, `for.*`, `while.*`, `switch.*`, `case.*`,
		`\.mw-parser-output.*`, `\.ba-charinfo.*`, `\.textToggleDisplay.*`,
		`@keyframes.*`, `animation.*`, `transform.*`, `transition.*`,
	}

	for _, pattern := range uselessPatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	// 分割成行并清理每一行
	lines := strings.Split(text, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 5 && !t.isUselessText(line) {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// 解码HTML实体
func (t *CrawlerTool) decodeHTMLEntities(text string) string {
	// 常见的HTML实体替换
	entities := map[string]string{
		"&#160;": " ", "&nbsp;": " ",
		"&#38;": "&", "&amp;": "&",
		"&#60;": "<", "&lt;": "<",
		"&#62;": ">", "&gt;": ">",
		"&#34;": "\"", "&quot;": "\"",
		"&#39;": "'", "&#039;": "'", "&apos;": "'",
		"&#169;": "©", "&copy;": "©",
	}

	for entity, replacement := range entities {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	// 处理数字实体（如 &#1234;）
	re := regexp.MustCompile(`&#(\d+);`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		return "" // 直接移除数字实体
	})

	return text
}

// 判断是否是无用文本
func (t *CrawlerTool) isUselessText(text string) bool {
	uselessKeywords := []string{
		"跳到导航", "跳到搜索", "编辑", "查看历史", "搜索",
		"导航", "首页", "最近更改", "随机页面", "帮助",
		"相关更改", "特殊页面", "打印版本", "永久链接",
		"页面信息", "引用本页", "本文介绍的是", "关于", "参见",
		"use strict", "window.RLQ.push", "function", "const",
	}

	for _, keyword := range uselessKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

// 爬取网页内容
func (t *CrawlerTool) crawl(url, extract string) (map[string]any, error) {
	// 获取页面内容
	htmlContent, err := t.FetchPage(url)
	if err != nil {
		return nil, err
	}

	// 提取内容
	content, err := t.ExtractCleanContent(htmlContent)
	if err != nil {
		return nil, err
	}

	// 提取标题
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}
	title := doc.Find("title").Text()
	title = strings.Replace(title, " - 萌娘百科", "", 1)
	title = strings.TrimSpace(title)

	// 构建结果
	result := make(map[string]any)

	switch extract {
	case "title":
		result["title"] = title
	case "content":
		result["content"] = content
	case "all":
		result["title"] = title
		result["content"] = content
		result["url"] = url
	}

	return result, nil
}

// 保存爬取结果到文件
func (t *CrawlerTool) saveResult(result map[string]any, filename, format string) error {
	switch format {
	case "txt":
		return t.saveAsTxt(result, filename)
	default:
		return fmt.Errorf("不支持的输出格式: %s", format)
	}
}

// 保存为文本文件
func (t *CrawlerTool) saveAsTxt(result map[string]any, filename string) error {
	var content strings.Builder

	// 写入标题
	if title, ok := result["title"].(string); ok {
		content.WriteString("标题: " + title + "\n\n")
	}

	// 写入 URL
	if url, ok := result["url"].(string); ok {
		content.WriteString("URL: " + url + "\n\n")
	}

	// 写入内容
	if contentText, ok := result["content"].(string); ok {
		content.WriteString("内容:\n" + contentText + "\n")
	}

	// 写入文件
	return os.WriteFile(filename, []byte(content.String()), 0644)
}
