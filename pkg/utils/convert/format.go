package convert

import "fmt"

func FormatFileSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}

	// 计算对数，确定单位等级
	// size = 1024 -> i = 1 (KB)
	i := 0
	fSize := float64(size)
	for fSize >= 1024 && i < len(units)-1 {
		fSize /= 1024
		i++
	}
	// 保留两位小数
	return fmt.Sprintf("%.2f %s", fSize, units[i])
}
