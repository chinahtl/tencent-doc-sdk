package util

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DownloadFromCOS 从腾讯云 COS（对象存储）下载文件到本地指定目录。
//
// 参数:
//   - fileURL: 文件的完整 URL 地址，指向腾讯云 COS 上的资源。
//   - saveDir: 文件保存的本地目录路径。如果为空，则默认保存到当前工作目录。
//
// 返回值:
//   - string: 下载文件的完整本地路径。
//   - error: 如果下载过程中发生错误，返回具体的错误信息；否则返回 nil。
//
// 示例:
//
//	filePath, err := DownloadFromCOS("https://example.com/file.txt", "./downloads")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("文件已下载到:", filePath)
//
// 功能说明:
//  1. 发起 HTTP GET 请求下载文件。
//  2. 检查响应状态码，确保请求成功。
//  3. 从响应头中提取文件名。
//  4. 创建本地目录（如果不存在）。
//  5. 将文件流式写入本地路径。
//  6. 返回下载文件的完整路径。
func DownloadFromCOS(fileURL, saveDir string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 获取文件名
	fileName, err := getFileName(resp)
	if err != nil {
		return "", fmt.Errorf("获取文件名失败: %w", err)
	}

	// 处理保存目录
	if saveDir == "" {
		saveDir = "." // 当前目录
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 构建完整保存路径
	fullPath := filepath.Join(saveDir, fileName)

	// 创建本地文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("文件创建失败: %w", err)
	}
	defer file.Close()

	// 流式写入文件
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("文件写入失败: %w", err)
	}

	return fullPath, nil
}

// getFileName 获取文件名
func getFileName(resp *http.Response) (string, error) {
	disposition := resp.Header.Get("Content-Disposition")
	if disposition == "" {
		return "", fmt.Errorf("Content-Disposition头不存在")
	}

	_, params, err := mime.ParseMediaType(disposition)
	if err == nil {
		if fileName := fileNameFromParams(params); fileName != "" {
			return sanitizeFileName(fileName), nil
		}
	}

	fileName := fileNameFromDisposition(disposition)
	if fileName == "" && err != nil {
		return "", fmt.Errorf("解析Content-Disposition失败: %w", err)
	}
	if fileName == "" {
		return "", fmt.Errorf("无法从响应头中提取文件名")
	}

	return sanitizeFileName(fileName), nil
}

func fileNameFromParams(params map[string]string) string {
	// 优先使用UTF-8编码的文件名
	if name := params["filename*"]; name != "" {
		if fileName := decodeRFC5987Value(name); fileName != "" {
			return fileName
		}
	}

	// Go 标准库有时会将 filename* 直接归一化到 filename。
	if name := params["filename"]; name != "" {
		return strings.Trim(name, `"`)
	}

	return ""
}

func fileNameFromDisposition(disposition string) string {
	if name := findHeaderParam(disposition, "filename*"); name != "" {
		if fileName := decodeRFC5987Value(name); fileName != "" {
			return fileName
		}
	}

	if name := findHeaderParam(disposition, "filename"); name != "" {
		return strings.Trim(name, `"`)
	}

	return ""
}

func findHeaderParam(disposition, key string) string {
	pattern := regexp.MustCompile(`(?i)(?:^|;)\s*` + regexp.QuoteMeta(key) + `=([^;]+)`)
	matches := pattern.FindStringSubmatch(disposition)
	if len(matches) < 2 {
		return ""
	}

	return strings.TrimSpace(matches[1])
}

func decodeRFC5987Value(value string) string {
	value = strings.Trim(value, `"`)
	if strings.Contains(value, "''") {
		parts := strings.SplitN(value, "''", 2)
		value = parts[1]
	}

	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return ""
	}

	return decoded
}

func sanitizeFileName(fileName string) string {
	// 替换文件名中的路径分隔符，防止创建文件时出错
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	return fileName
}
