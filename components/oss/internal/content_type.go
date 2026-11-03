package internal

import (
	"path"
	"strings"
)

// contentTypeByObjectKey 按对象键扩展名推断 Content-Type。PutObject 此前不设
// Content-Type，RustFS/R2/MinIO/Kodo 均落成 application/octet-stream：浏览器在
// X-Content-Type-Options: nosniff 下不渲染图片。这里用显式小表而非
// mime.TypeByExtension——后者依赖宿主系统 mime.types/注册表，容器内行为不稳定，
// 且常见类型集合本就有限。未知扩展名回退 application/octet-stream（与原行为一致）。
// 注意：不得用于用户可控对象键（.html 会落 text/html）。
func contentTypeByObjectKey(objectKey string) string {
	switch strings.ToLower(path.Ext(objectKey)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt", ".log":
		return "text/plain; charset=utf-8"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".zip":
		return "application/zip"
	case ".gz":
		return "application/gzip"
	case ".tar":
		return "application/x-tar"
	case ".apk":
		return "application/vnd.android.package-archive"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	default:
		return "application/octet-stream"
	}
}
