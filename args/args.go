package args

import (
	COM "jpm/common"
	"os"
)

func Args(key string) string {
	COM.FindPackageYML(false)
	args := COM.GetSection("args", true).(map[string]string)
	if v, exists := args[key]; exists {
		return v
	}
	os.Exit(404)
	return ""
}
