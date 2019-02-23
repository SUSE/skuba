package salt

import (
	"fmt"
)

func SaltPath(path string) string {
	return fmt.Sprintf("salt://%s", path)
}
